package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"
)

type Config struct {
	Pattern    string
	Workers    int
	Anywhere   bool
	IgnoreCase bool
	Stream     bool
	Verbose    bool
}

func main() {
	var (
		pattern    = flag.String("pattern", "", "Pattern to match in public key")
		workers    = flag.Int("workers", runtime.NumCPU(), "Number of CPU workers")
		anywhere   = flag.Bool("anywhere", true, "Search anywhere in key (not just start)")
		ignoreCase = flag.Bool("ignore-case", true, "Case insensitive matching")
		stream     = flag.Bool("stream", false, "Keep finding matches (streaming mode)")
		verbose    = flag.Bool("verbose", false, "verbose logging")
	)
	flag.Parse()

	if *pattern == "" {
		fmt.Println("Sleipnir - Vanity SSH Key Generator")
		fmt.Println("\nUsage: sleipnir -pattern <string>")
		fmt.Println("\nExamples:")
		fmt.Println("  sleipnir -pattern cool        # Find 'cool' anywhere in key")
		fmt.Println("  sleipnir -pattern 'l33t|elite' # Multiple patterns")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("Sleipnir galloping with %d workers...\n", *workers)
	fmt.Printf("Hunting pattern: %v\n", *pattern)
	fmt.Println("Press Ctrl+C to stop")

	config := &Config{
		Pattern:    *pattern,
		Workers:    *workers,
		Anywhere:   *anywhere,
		IgnoreCase: *ignoreCase,
		Stream:     *stream,
		Verbose:    *verbose,
	}

	var wg sync.WaitGroup
	result := startGen(config, &wg)
	fmt.Printf("\nKEY FOUND :)!\nPrivate Key:\n%v\nPublic Key:\n%v\n", result.priv, result.pub)
	wg.Wait()
	fmt.Println("All goroutines closed successfully")

}
