package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"
	"strings"

	cl "github.com/CyberChainXyz/go-opencl"
)

type GPUContext struct {
	runner    *cl.OpenCLRunner
	batchSize int

	seedBuffer    *cl.Buffer
	pubBuffer     *cl.Buffer
	privBuffer    *cl.Buffer
	countBuffer   *cl.Buffer
	patternBuffer *cl.Buffer

	workGroupSize  uint64
	globalWorkSize uint64
}

func locationToInt(location string) int {
	switch location {
	case "anywhere":
		return 0
	case "start":
		return 1
	case "end":
		return 2
	default:
		return 2
	}
}

func initGpu(config *Config) (*GPUContext, error) {
	pattern := config.Patterns[0]
	patternBytes := []byte(pattern)

	info, err := cl.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenCL info: %v", err)
	}

	if len(info.Platforms) == 0 || len(info.Platforms[0].Devices) == 0 {
		return nil, fmt.Errorf("no OpenCL devices found")
	}

	device := info.Platforms[0].Devices[0]
	runner, err := device.InitRunner()
	if err != nil {
		return nil, fmt.Errorf("failed to init runner: %v", err)
	}
	// Skip cleanup to avoid segfault for now

	// Load kernel
	kernelSource, err := loadSleipnirKernel()
	if err != nil {
		return nil, fmt.Errorf("failed to load kernel: %v", err)
	}

	err = runner.CompileKernels([]string{kernelSource}, []string{"sleipnir_ed25519_keygen"}, "")
	if err != nil {
		return nil, fmt.Errorf("kernel compilation failed: %v", err)
	}

	batchSize := config.BatchSize

	// Create GPU buffers
	seedBuffer, err := runner.CreateEmptyBuffer(cl.READ_ONLY, batchSize*32)
	if err != nil {
		return nil, fmt.Errorf("failed to create seed buffer: %v", err)
	}

	publicKeyBuffer, err := runner.CreateEmptyBuffer(cl.WRITE_ONLY, batchSize*32)
	if err != nil {
		return nil, fmt.Errorf("failed to create public key buffer: %v", err)
	}

	privateKeyBuffer, err := runner.CreateEmptyBuffer(cl.WRITE_ONLY, batchSize*64) // ADD: 64 bytes per key
	if err != nil {
		return nil, fmt.Errorf("failed to create private key buffer: %v", err)
	}

	countBuffer, err := runner.CreateEmptyBuffer(cl.READ_WRITE, 4)
	if err != nil {
		return nil, fmt.Errorf("failed to create count buffer: %v", err)
	}

	patternBuffer, err := runner.CreateEmptyBuffer(cl.READ_ONLY, len(patternBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create pattern buffer: %v", err)
	}

	workGroupSize := uint64(256)
	globalWorkSize := uint64((batchSize + 255) / 256 * 256)

	return &GPUContext{
		runner:         runner,
		batchSize:      batchSize,
		seedBuffer:     seedBuffer,
		pubBuffer:      publicKeyBuffer,
		privBuffer:     privateKeyBuffer,
		countBuffer:    countBuffer,
		patternBuffer:  patternBuffer,
		workGroupSize:  workGroupSize,
		globalWorkSize: globalWorkSize,
	}, nil
}

func (g *GPUContext) findVanityKeysGPU(config *Config) (*resultFound, error) {
	pattern := config.Patterns[0]
	if config.IgnoreCase {
		pattern = strings.ToLower(pattern)
	}
	patternBytes := []byte(pattern)
	// Take first pattern only for testing now CHANGE THIS
	patternLen := int32(len(pattern))
	locationInt := int32(locationToInt(config.Location))
	ignoreCaseInt := int32(0)
	if config.IgnoreCase {
		ignoreCaseInt = 1
	}

	// Generate random seeds, we do this in go and not C so we can make it crypto secure
	seeds := make([]byte, g.batchSize*32)
	_, err := rand.Read(seeds)
	if err != nil {
		return nil, err
	}

	// Write buffers to gpu
	err = cl.WriteBuffer(g.runner, 0, g.seedBuffer, seeds, true)
	if err != nil {
		return nil, fmt.Errorf("failed to write seeds: %v", err)
	}

	matchCount := []int32{0}
	err = cl.WriteBuffer(g.runner, 0, g.countBuffer, matchCount, true)
	if err != nil {
		return nil, fmt.Errorf("failed to write count: %v", err)
	}

	err = cl.WriteBuffer(g.runner, 0, g.patternBuffer, patternBytes, true)
	if err != nil {
		return nil, fmt.Errorf("failed to write pattern: %v", err)
	}

	// Run kernel
	batchSizeParam := int32(g.batchSize)
	args := []cl.KernelParam{
		cl.BufferParam(g.seedBuffer),
		cl.BufferParam(g.pubBuffer),
		cl.BufferParam(g.privBuffer),
		cl.BufferParam(g.countBuffer),
		cl.BufferParam(g.patternBuffer),
		cl.Param(&batchSizeParam),
		cl.Param(&patternLen),
		cl.Param(&locationInt),
		cl.Param(&ignoreCaseInt),
	}

	workGroupSize := uint64(256)                              // Check for config
	globalWorkSize := uint64((g.batchSize + 255) / 256 * 256) // Round up

	err = g.runner.RunKernel("sleipnir_ed25519_keygen", 1, nil, []uint64{globalWorkSize}, []uint64{workGroupSize}, args, true)
	if err != nil {
		return nil, fmt.Errorf("kernel execution failed: %v", err)
	}

	// Read results
	err = cl.ReadBuffer(g.runner, 0, g.countBuffer, matchCount)
	if err != nil {
		return nil, fmt.Errorf("failed to read count: %v", err)
	}

	if matchCount[0] == 0 {
		return nil, nil
	}

	// Read the first match, fix this when we start streaming
	foundPrivateKeys := make([]byte, 128)
	err = cl.ReadBuffer(g.runner, 0, g.privBuffer, foundPrivateKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to read indices: %v", err)
	}

	foundPublicKeys := make([]byte, 80)
	err = cl.ReadBuffer(g.runner, 0, g.pubBuffer, foundPublicKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to read public keys: %v", err)
	}

	data, err := privateKeyToOpenSSH(ed25519.PrivateKey(foundPrivateKeys), "")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	privKey, err := privateKeyToPEM(ed25519.PrivateKey(foundPrivateKeys))
	if err != nil {
		return nil, err
	}

	return &resultFound{
		pub:         string(foundPublicKeys),
		priv:        privKey,
		privOpenSSH: data,
	}, nil
}

func loadSleipnirKernel() (string, error) {
	kernelBytes, err := os.ReadFile("gpu/sleipnir_kernel.cl")
	if err != nil {
		return "", fmt.Errorf("failed to read kernel file: %v", err)
	}
	return string(kernelBytes), nil
}
