package main

import (
	"fmt"
	"math"
	"os"
)

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

func writeKey(result *resultFound, cfg *Config) {
	f, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		err = f.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}()

	newLine := fmt.Sprintf("\n-------------------\nKEY FOUND :)!\nOpenSSH Private Key:\n%v\nPublic Key:\nssh-ed25519 %v\n-------------------\n", result.privOpenSSH, result.pub)
	if cfg.Verbose {
		newLine += fmt.Sprintf("PKCS#8 Private Key:\n%v\n", result.priv)
	}
	_, err = fmt.Fprintln(f, newLine)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
}
