# Sleipnir - Super Fast Vanity SSH Key Generator

Sleipnir is a cross-platform vanity SSH key generator written in Go, capable of generating **26,000,000+** ED25519 keys per second on modern hardware using both CPU and GPU processing. Named after Odin's eight-legged horse from Norse mythology, Sleipnir gallops through keyspace at incredible speeds to find your perfect vanity SSH key.

## Compiling
See [docs/compiling.md](docs/compiling.md) for Linux, Windows, and cross-compilation instructions.

## Usage

```bash
# Find "cool" anywhere in the SSH key
./sleipnir -pattern cool

# Find a key ending with "1337", "meng", or "github"
./sleipnir -pattern 1337,meng,github -location end

# GPU-only mode (recommended when using a GPU — see note below)
./sleipnir -pattern mari -location end -gpu -cpu=false -batch-size 33554432
```

For more examples see [docs/usage.md](docs/usage.md).

> [!TIP]
> **When using a GPU, disable the CPU workers** (`-cpu=false`). A modern GPU is 20–30× faster
> than the CPU, so running CPU workers alongside it gives negligible extra throughput — and the
> CPU still needs headroom to feed the GPU with fresh random seeds. Tune `-batch-size` to find
> the sweet spot for your card (double or halve the default until keys/s peaks).

> [!NOTE]
> AMD GPU support requires the ROCm or AMDGPU-PRO OpenCL runtime.
> See [docs/compiling.md](docs/compiling.md) for installation instructions.

## Example

```bash
./sleipnir -pattern MENG -location end -ignore-case=false
```
```
Sleipnir galloping with 24 workers...
Hunting pattern: MENG
Press Ctrl+C to stop
Expected tries: 1.6777216e+07
|Average keys per second: 977558| |Total tries: 4887855| |Calculated wait time: 0d 00h 00m 05s/0d 00h 00m 17s|
|Average keys per second: 975424| |Total tries: 9764953| |Calculated wait time: 0d 00h 00m 10s/0d 00h 00m 17s|
...
|Average keys per second: 979482| |Total tries: 29326450| |Calculated wait time: 0d 00h 00m 30s/0d 00h 00m 17s|
Made it in 33054311 tries

KEY FOUND :)!
OpenSSH Private Key:
-----BEGIN OPENSSH PRIVATE KEY-----
Removed so nobody would actually use this key :p
-----END OPENSSH PRIVATE KEY-----

Public Key:
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAII44C87jrgvZi/pkNUVpwb0jlnUGXkiUu+/RMS5wMENG
All goroutines closed successfully
```

> [!NOTE]
> Add `-verbose` to also print the PKCS#8 format private key.

## Tests

Benchmark raw key generation speed and overall Sleipnir throughput:
```bash
go test -bench .
```

Test that generated keys are valid:
```bash
go test -v
```

### Flamegraph

Run Sleipnir with `-pprof`, then:
```bash
go tool pprof -http=:8000 sleipnir.pprof
```
Visit http://localhost:8000/ui/flamegraph

## Benchmarks

| GPU                     | keys/s | OS              | Version   | batch-size |
|:------------------------|:------:|:----------------|-----------|------------|
| GeForce RTX 4070 12GB   | ~26M   | Debian Linux 13 | 1.0.0     | 33554432   |
| GeForce RTX 3080 10GB   | ~23M   | Arch Linux      | 1.0.0     | 16777216   |
| GeForce RTX 3080 10GB   | ~17M   | Windows 11 23H2 | pre-1.0.0 | ?          |
| GeForce RTX 3060 Ti 8GB | ~12.5M | Windows 11 23H2 | pre-1.0.0 | ?          |
| GeForce RTX 4070 12GB   | ~9.5M  | Debian Linux 12 | pre-1.0.0 | ?          |

| CPU                  | keys/s | OS              | Version   |
|:---------------------|:------:|:----------------|-----------|
| Intel Core i7-13700K | ~1M    | Windows 11 23H2 | pre-1.0.0 |
| AMD Ryzen 9 7950X    | ~920k  | Fedora Linux 42 | pre-1.0.0 |
| AMD Ryzen 7 7800x3d  | ~570K  | Fedora Linux 42 | pre-1.0.0 |
| AMD Ryzen 5 7600X    | ~500K  | Debian Linux 12 | pre-1.0.0 |
| Apple M1             | ~280k  | macOS 26        | pre-1.0.0 |
| lx2160a A72          | ~143K  | Fedora Linux 42 | pre-1.0.0 |

Have a result to share? Open a PR or issue!
