package main

import (
	"bytes"
	"encoding/pem"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func BenchmarkEd25519Keygen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = generateED25519Key()
	}
}

func BenchmarkSleipnirThroughput(b *testing.B) {
	workers := runtime.NumCPU()
	cfg := &Config{
		Workers:    workers,
		Patterns:   []string{"ImagineSomeoneFindsThis"}, // "impossible" match
		Location:   "anywhere",
		IgnoreCase: true,
	}

	var wg sync.WaitGroup
	var total uint64

	// start timer after setup
	b.ResetTimer()
	start := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, len(sshEd25519Prefix)+32)
			for j := 0; j < b.N/workers; j++ {
				priv, pub, err := generateED25519Key()
				if err != nil {
					return
				}

				pubString := publicKeyToSSHFormat(pub, buf)
				if checkKey(pubString, cfg) {
					_, err := privateKeyToPEM(priv)
					if err != nil {
						return
					}
					_, privErr := privateKeyToOpenSSH(priv, "")
					if privErr != nil {
						return
					}
				}
				atomic.AddUint64(&total, 1)
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start).Seconds()
	keysPerSec := float64(total) / duration

	b.ReportMetric(keysPerSec, "keys/s")
}

func TestPublicKeyFormat(t *testing.T) {
	buf := make([]byte, len(sshEd25519Prefix)+32)
	_, pub, err := generateED25519Key()
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	pubStr := publicKeyToSSHFormat(pub, buf)
	keyData := []byte("ssh-ed25519 " + pubStr)
	_, _, _, _, errFrom := ssh.ParseAuthorizedKey(keyData)
	if errFrom != nil {
		t.Fatalf("public key is invalid: %v", err)
	}
}

func TestPrivateKeyPEM(t *testing.T) {
	priv, _, err := generateED25519Key()
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	pemStr, err := privateKeyToPEM(priv)
	if err != nil {
		t.Fatalf("failed to encode PEM: %v", err)
	}

	block, _ := pem.Decode([]byte(pemStr))
	if block == nil || !bytes.HasPrefix(block.Bytes, []byte{0x30}) {
		t.Fatalf("PEM block is invalid")
	}
}
