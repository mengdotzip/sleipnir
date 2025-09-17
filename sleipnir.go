package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
)

type Config struct {
	Pattern    string
	Workers    int
	Location   string
	IgnoreCase bool
	Stream     bool
	Verbose    bool
}

func main() {
	var (
		pattern    = flag.String("pattern", "", "Pattern to match in public key")
		workers    = flag.Int("workers", runtime.NumCPU(), "Number of CPU workers")
		location   = flag.String("location", "anywhere", "Search 'anywhere/start/end' of the public key")
		ignoreCase = flag.Bool("ignore-case", true, "Case insensitive matching")
		stream     = flag.Bool("stream", false, "Keep finding matches (streaming mode)")
		verbose    = flag.Bool("verbose", false, "verbose logging")
	)
	flag.Parse()

	if *pattern == "" {
		fmt.Println("Sleipnir - Vanity SSH Key Generator")
		fmt.Println("\nUsage: sleipnir -pattern <string>")
		fmt.Println("\nExamples:")
		fmt.Println("sleipnir -pattern cool -location anywhere  # Find 'cool' anywhere in key")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *location == "" {
		fmt.Println("UNKOWN: Use '-location start,end or anywhere'")
		os.Exit(1)
	}

	fmt.Printf("Sleipnir galloping with %d workers...\n", *workers)
	fmt.Printf("Hunting pattern: %v\n", *pattern)
	fmt.Println("Press Ctrl+C to stop")

	config := &Config{
		Pattern:    *pattern,
		Workers:    *workers,
		Location:   *location,
		IgnoreCase: *ignoreCase,
		Stream:     *stream,
		Verbose:    *verbose,
	}

	if config.IgnoreCase {
		config.Pattern = strings.ToLower(config.Pattern)
	}

	var wg sync.WaitGroup
	result := startGen(config, &wg)
	if result != nil {
		fmt.Printf("\nKEY FOUND :)!\n")
		if config.Verbose {
			fmt.Printf("PKCS#8 Private Key:\n%v\n", result.priv)
		}
		fmt.Printf("OpenSSH Private Key:\n%v\nPublic Key:\n%v\n", result.privOpenSSH, result.pub)
	}
	wg.Wait()
	fmt.Println("All goroutines closed successfully")

}
