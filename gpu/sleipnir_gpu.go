package gpu

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"
	"time"

	cl "github.com/CyberChainXyz/go-opencl"
)

type GPUResult struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
	SSHKey     string
}

type GPUConfig struct {
	Patterns   []string
	Location   string
	IgnoreCase bool
	BatchSize  int // Keys to generate per GPU call
}

func FindVanityKeysGPU(patterns []string, location string, ignoreCase bool) (*GPUResult, error) {
	config := GPUConfig{
		Patterns:   patterns,
		Location:   location,
		IgnoreCase: ignoreCase,
		BatchSize:  16340, // Manual config for now, change this!
	}
	// Get OpenCL device
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
	kernelSource, err := loadSleipnirKernel(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load kernel: %v", err)
	}

	err = runner.CompileKernels([]string{kernelSource}, []string{"sleipnir_ed25519_keygen"}, "")
	if err != nil {
		return nil, fmt.Errorf("kernel compilation failed: %v", err)
	}

	start := time.Now()
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

	seedIndexBuffer, err := runner.CreateEmptyBuffer(cl.WRITE_ONLY, batchSize*4)
	if err != nil {
		return nil, fmt.Errorf("failed to create index buffer: %v", err)
	}

	countBuffer, err := runner.CreateEmptyBuffer(cl.READ_WRITE, 4)
	if err != nil {
		return nil, fmt.Errorf("failed to create count buffer: %v", err)
	}

	// Generate random seeds
	seeds := make([]byte, batchSize*32)
	_, err = rand.Read(seeds)
	if err != nil {
		return nil, err
	}

	// Write data to GPU
	err = cl.WriteBuffer(runner, 0, seedBuffer, seeds, true)
	if err != nil {
		return nil, fmt.Errorf("failed to write seeds: %v", err)
	}

	// Initialize match count
	matchCount := []int32{0}
	err = cl.WriteBuffer(runner, 0, countBuffer, matchCount, true)
	if err != nil {
		return nil, fmt.Errorf("failed to write count: %v", err)
	}

	// Run kernel
	batchSizeParam := int32(batchSize)
	args := []cl.KernelParam{
		cl.BufferParam(seedBuffer),
		cl.BufferParam(publicKeyBuffer),
		cl.BufferParam(seedIndexBuffer),
		cl.BufferParam(countBuffer),
		cl.Param(&batchSizeParam),
	}

	workGroupSize := uint64(256)                            // Check for config
	globalWorkSize := uint64((batchSize + 255) / 256 * 256) // Round up

	err = runner.RunKernel("sleipnir_ed25519_keygen", 1, nil, []uint64{globalWorkSize}, []uint64{workGroupSize}, args, true)
	if err != nil {
		return nil, fmt.Errorf("kernel execution failed: %v", err)
	}

	elapsed := time.Since(start)
	keysPerSecond := float64(batchSize) / elapsed.Seconds()

	// Read results
	err = cl.ReadBuffer(runner, 0, countBuffer, matchCount)
	if err != nil {
		return nil, fmt.Errorf("failed to read count: %v", err)
	}

	fmt.Printf("GPU Performance: %.0f keys/sec (batch: %d, time: %v)\n",
		keysPerSecond, batchSize, elapsed)
	fmt.Printf("Found %d matches out of %d keys!\n", matchCount[0], batchSize)

	if matchCount[0] == 0 {
		return nil, nil // No matches found
	}

	// Read the first match
	foundSeedIndices := make([]int32, 1)
	err = cl.ReadBuffer(runner, 0, seedIndexBuffer, foundSeedIndices)
	if err != nil {
		return nil, fmt.Errorf("failed to read indices: %v", err)
	}

	foundPublicKeys := make([]byte, 32)
	err = cl.ReadBuffer(runner, 0, publicKeyBuffer, foundPublicKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to read public keys: %v", err)
	}

	// Extract the matching seed and regenerate private key
	seedIndex := foundSeedIndices[0]
	matchingSeed := make([]byte, 32)
	copy(matchingSeed, seeds[seedIndex*32:(seedIndex+1)*32])

	// Generate the full keypair from the seed
	privateKey := ed25519.NewKeyFromSeed(matchingSeed)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	// Format as SSH key
	sshKey := formatSSHPublicKey(publicKey)

	return &GPUResult{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		SSHKey:     sshKey,
	}, nil
}

func loadSleipnirKernel(config GPUConfig) (string, error) {
	kernelBytes, err := os.ReadFile("gpu/sleipnir_kernel.cl")
	if err != nil {
		return "", fmt.Errorf("failed to read kernel file: %v", err)
	}
	return string(kernelBytes), nil
}

// Format ED25519 public key as SSH public key (test for now, use actual one)
func formatSSHPublicKey(pubKey ed25519.PublicKey) string {

	keyBytes := []byte(pubKey)

	return fmt.Sprintf("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI%s", string(keyBytes))
}
