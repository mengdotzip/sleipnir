package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var tries uint64 = 0

type resultFound struct {
	pub  string
	priv string
}

//SSH GEN

func generateED25519Key() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return priv, pub, nil
}

var sshEd25519Prefix = []byte{
	0, 0, 0, 11, // length of "ssh-ed25519"
	's', 's', 'h', '-', 'e', 'd', '2', '5', '5', '1', '9',
	0, 0, 0, 32, // length of pubkey
}

func publicKeyToSSHFormat(pub ed25519.PublicKey, buf []byte) string {
	n := copy(buf, sshEd25519Prefix)
	n += copy(buf[n:], pub)

	return base64.RawStdEncoding.EncodeToString(buf[:n])
}

func privateKeyToPEM(priv ed25519.PrivateKey) (string, error) {
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", err
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	}

	return string(pem.EncodeToMemory(block)), nil
}

//FUNCS

func checkKey(pub string, cfg *Config) bool {
	if cfg.IgnoreCase {
		pub = strings.ToLower(pub)
	}

	switch cfg.Location {
	case "anywhere":
		if strings.Contains(pub, cfg.Pattern) {
			return true
		}
	case "start":
		if strings.HasPrefix(pub, cfg.Pattern) {
			return true
		}
	case "end":
		if strings.HasSuffix(pub, cfg.Pattern) {
			return true
		}
	default:
		fmt.Println("Only use 'anywhere,start or end' as location flags!")
	}

	return false
}

func cpuGen(ctx context.Context, cfg *Config, result chan *resultFound, wg *sync.WaitGroup) {
	defer wg.Done()

	//each worker has 1 buffer (instead of making a new one every time)
	buf := make([]byte, len(sshEd25519Prefix)+32)

	for {
		select {
		case <-ctx.Done():
			if cfg.Verbose {
				fmt.Println("stopping cpu loop")
			}
			return
		default:
			priv, pub, err := generateED25519Key()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			pubString := publicKeyToSSHFormat(pub, buf)
			if checkKey(pubString, cfg) {
				privString, err := privateKeyToPEM(priv)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				fmt.Printf("Made it in %v tries\n", tries)
				result <- &resultFound{pubString, privString}
				return
			}

			//I wanted to use atomic.AddUint64(&tries, 1) but that 2ns overhead just too much
			atomic.AddUint64(&tries, 1)
		}
	}
}

func formatSeconds(sec float64) string {
	if math.IsInf(sec, 0) {
		return "∞"
	}

	days := int(sec) / 86400
	sec -= float64(days * 86400)
	hours := int(sec) / 3600
	sec -= float64(hours * 3600)
	mins := int(sec) / 60
	sec -= float64(mins * 60)
	return fmt.Sprintf("%dd %02dh %02dm %02ds", days, hours, mins, int(sec))
}

// I asked the ai for the formula, feel free to optimize this
func estimateTries(pattern string, location string, ignoreCase bool) float64 {
	const keyLen int = 43

	n := len(pattern)
	if n == 0 || keyLen < n {
		return math.Inf(1)
	}

	charset := 64 // base64
	if ignoreCase {
		// roughly half the charset
		charset = 32
	}

	// probability single position matches
	p := math.Pow(1.0/float64(charset), float64(n))

	positions := 1
	switch location {
	case "start", "end":
		positions = 1
	case "anywhere":
		positions = keyLen - n + 1
		if positions < 1 {
			positions = 1
		}
	default:
		positions = 1
	}

	// expected tries = 1 / (p * positions)
	return 1.0 / (p * float64(positions))
}

func stats(ctx context.Context, cfg *Config) {
	ticker := time.NewTicker(5 * time.Second)
	start := time.Now()
	defer ticker.Stop()
	var oldTries uint64 = 0

	expectedTries := estimateTries(cfg.Pattern, cfg.Location, cfg.IgnoreCase)
	fmt.Printf("Expected tries: %v\n", expectedTries)

	for {
		select {
		case <-ctx.Done():
			if cfg.Verbose {
				fmt.Println("stopping stats loop")
			}
			return
		case <-ticker.C:
			deltaT := tries - oldTries
			oldTries = tries
			keysPS := deltaT / 5
			etaSec := float64(expectedTries) / float64(keysPS)
			fmt.Printf("Average keys per second: %v Total tries: %v Calculated wait time: %v/%v\n", keysPS, tries, formatSeconds(time.Since(start).Seconds()), formatSeconds(etaSec))
		}
	}
}

func startGen(cfg *Config, wg *sync.WaitGroup) *resultFound {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	result := make(chan *resultFound, 1)
	go stats(ctx, cfg)

	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go cpuGen(ctx, cfg, result, wg)
	}

	select {
	case foundResult := <-result:
		stop()
		return foundResult
	case <-ctx.Done():
		return nil
	}

}
