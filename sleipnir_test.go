package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func BenchmarkEd25519Keygen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = ed25519.GenerateKey(rand.Reader)
	}
}
