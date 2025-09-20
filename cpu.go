package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

func cpuGen(ctx context.Context, cfg *Config, result chan *resultFound, wg *sync.WaitGroup) {
	defer wg.Done()

	// each worker has 1 buffer (instead of making a new one every time)
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
				ctx.Done()
				os.Exit(1)
			}

			pubString := publicKeyToSSHFormat(pub, buf)
			if checkKey(pubString, cfg) {
				privString, err := privateKeyToPEM(priv)
				if err != nil {
					fmt.Println(err)
					ctx.Done()
					os.Exit(1)
				}
				privOpenSSH, err := privateKeyToOpenSSH(priv, "")
				if err != nil {
					fmt.Println(err)
					ctx.Done()
					os.Exit(1)
				}

				fmt.Printf("Made it in %v tries\n", tries)
				result <- &resultFound{pubString, privString, privOpenSSH}
				return
			}

			atomic.AddUint64(&tries, 1)
		}
	}
}
