package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
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
	useCpu     bool
	UseGpu     bool
	BatchSize  int
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
		useCpu     = flag.Bool("cpu", true, "Use your CPU for generation (slower)")
		useGpu     = flag.Bool("gpu", false, "Use your GPU for generation")
		batchSize  = flag.Int("batch-size", 65536, "Amount of workers per gpu call")
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
		useCpu:     *useCpu,
		UseGpu:     *useGpu,
		BatchSize:  *batchSize,
	}
	//------------

	fmt.Printf("Sleipnir galloping with %d workers...\n", *workers)
	fmt.Printf("Hunting pattern: %v\n", *pattern)
	fmt.Println("Press Ctrl+C to stop")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var wg sync.WaitGroup
	go stats(ctx, config)
	if config.UseGpu {
		fmt.Println("WARNING: GPU is still in TESTING only -location end.anywhere works")
		wg.Add(1)
		go startGpuGen(config, &wg, ctx, stop)
	}

	if config.useCpu {
		if !config.Stream {
			foundResult := startGen(config, &wg, ctx, stop)
			printResult(foundResult, config)
			if config.Output != "" {
				writeKey(foundResult, config)
			}
		} else {
			startGenStream(config, &wg, ctx)
		}
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
		fmt.Printf("OpenSSH Private Key:\n%v\nPublic Key:\n%v\n", result.privOpenSSH, result.pub)
	}
}

func startGpuGen(config *Config, wg *sync.WaitGroup, ctx context.Context, stop context.CancelFunc) {
	defer wg.Done()
	gpuCTX, err := initGpu(config)
	if err != nil {
		fmt.Printf("Error loading the kernel %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			found, err := gpuCTX.findVanityKeysGPU(config)
			if err != nil {
				fmt.Println(err)
				stop()
				return
			}
			atomic.AddUint64(&tries, uint64(config.BatchSize))
			if found != nil {
				printResult(found, config)
				stop()
				return
			}
		}

	}
}

func startGenStream(cfg *Config, wg *sync.WaitGroup, ctx context.Context) {
	result := make(chan *resultFound, 1)

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

func startGen(cfg *Config, wg *sync.WaitGroup, ctx context.Context, stop context.CancelFunc) *resultFound {
	result := make(chan *resultFound, 1)

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
