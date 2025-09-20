package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sleipnir/gpu"
	"strings"
	"sync"
	"syscall"
)

type Config struct {
	Patterns   []string
	Workers    int
	Location   string
	IgnoreCase bool
	Stream     bool
	Output     string
	Verbose    bool
	UseGpu     bool
}

func main() {

	//Config creation
	var (
		pattern    = flag.String("pattern", "", "Pattern(s) to match in public key")
		workers    = flag.Int("workers", runtime.NumCPU(), "Number of CPU workers")
		location   = flag.String("location", "anywhere", "Search 'anywhere/start/end' of the public key")
		ignoreCase = flag.Bool("ignore-case", true, "Case insensitive matching")
		stream     = flag.Bool("stream", false, "Keep finding matches (streaming mode)")
		output     = flag.String("output", "", "Write found keys to a file")
		verbose    = flag.Bool("verbose", false, "verbose logging")
		useGpu     = flag.Bool("gpu", false, "Use your GPU for generation")
	)
	flag.Parse()

	if *pattern == "" {
		fmt.Println("Sleipnir - Vanity SSH Key Generator")
		fmt.Println("\nUsage: sleipnir -pattern <string>")
		fmt.Println("\nExamples:")
		fmt.Println("sleipnir -pattern cool -location anywhere  # Find 'cool' anywhere in key")
		fmt.Println("sleipnir -pattern cool,MENG -location end  # Find 'cool'or'MENG' at the end of the key")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *location == "" {
		// Technically we should never enter this function, but you never know.
		fmt.Println("UNKOWN: Use '-location start,end or anywhere'")
		os.Exit(1)
	}

	if *stream && *output == "" {
		fmt.Println("WARNING: I strongly suggest using -output to write the keys to a file while using stream.")
	}

	if *ignoreCase {
		*pattern = strings.ToLower(*pattern)
	}
	patterns := strings.Split(*pattern, ",")

	config := &Config{
		Patterns:   patterns,
		Workers:    *workers,
		Location:   *location,
		IgnoreCase: *ignoreCase,
		Stream:     *stream,
		Output:     *output,
		Verbose:    *verbose,
		UseGpu:     *useGpu,
	}
	//------------

	fmt.Printf("Sleipnir galloping with %d workers...\n", *workers)
	fmt.Printf("Hunting pattern: %v\n", *pattern)
	fmt.Println("Press Ctrl+C to stop")

	var wg sync.WaitGroup
	if config.UseGpu {
		fmt.Println("Testing OpenCL setup...")

		found, err := gpu.FindVanityKeysGPU(config.Patterns, config.Location, config.IgnoreCase)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("GPU found key:%v\n", found.PublicKey)
		buf := make([]byte, len(sshEd25519Prefix)+32)
		fmt.Printf("GPU found key:%v\n", publicKeyToSSHFormat(found.PublicKey, buf))

		return
	}

	if !config.Stream {
		foundResult := startGen(config, &wg)
		printResult(foundResult, config)
		if config.Output != "" {
			writeKey(foundResult, config)
		}
	} else {
		startGenStream(config, &wg)
	}

	wg.Wait()
	fmt.Println("All goroutines closed successfully")

}

func printResult(result *resultFound, cfg *Config) {
	if result != nil {
		fmt.Printf("\nKEY FOUND :)!\n")
		if cfg.Verbose {
			fmt.Printf("PKCS#8 Private Key:\n%v\n", result.priv)
		}
		fmt.Printf("OpenSSH Private Key:\n%v\nPublic Key:\nssh-ed25519 %v\n", result.privOpenSSH, result.pub)
	}
}

func startGenStream(cfg *Config, wg *sync.WaitGroup) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	result := make(chan *resultFound, 1)
	go stats(ctx, cfg)

	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go cpuGen(ctx, cfg, result, wg)
	}
	for {
		select {
		case foundResult := <-result:
			printResult(foundResult, cfg)
			if cfg.Output != "" {
				writeKey(foundResult, cfg)
			}
		case <-ctx.Done():
			return
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
