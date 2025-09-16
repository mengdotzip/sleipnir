package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
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

func publicKeyToSSHFormat(pub ed25519.PublicKey) string {
	keyData := make([]byte, 0, 51)

	algName := "ssh-ed25519"
	keyData = append(keyData, 0, 0, 0, 11)
	keyData = append(keyData, []byte(algName)...)

	keyData = append(keyData, 0, 0, 0, 32)
	keyData = append(keyData, pub...)

	return base64.StdEncoding.EncodeToString(keyData)
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

			pubString := publicKeyToSSHFormat(pub)
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

			tries++
		}
	}
}

func stats(ctx context.Context, cfg *Config) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	var oldTries uint64 = 0

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
			fmt.Printf("Average keys per second: %v Total tries: %v\n", keysPS, tries)
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
