# Sleipnir - Super Fast Vanity SSH Key Generator

Sleipnir is a super fast cross-platform vanity SSH key generator written in Go, capable of generating **17,000,000+** ED25519 keys per second on modern hardware using both CPU and GPU processing. Named after Odin's eight-legged horse from Norse mythology, Sleipnir gallops through keyspace at incredible speeds to find your perfect vanity SSH keys.

## Compiling
Please checkout the [DOCS](docs/compiling.md) for information on windows and Linux compiles.

## Usage

Basic Usage
```bash
# Find "cool" anywhere in the SSH key
./sleipnir -pattern cool

# Find key ending with "1337" OR "meng" OR "github"
./sleipnir -pattern 1337,meng,github -location end

# Use the gpu and cpu to find keys
./sleipnir -pattern mari -location end -gpu
```

For **more** usage examples please go to the [DOCS](docs/usage.md)


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
**NOTE** If you want the PKCS#8  format instead of OpenSSH you will have to run sleipnir with -verbose

## Tests
Benchmark the speed of the ssh keygen per core and the Sleipnir keys/s:
```
go test -bench .
```

Test if we are generating valid ssh keys:
```
go test -v
```

## Benchmarks

| GPU                     | keys/s | OS              |
|:------------------------|:------:|:----------------|
| GeForce RTX 3080 10GB   | ~17M   | Windows 11 23H2 |
| GeForce RTX 3060 Ti 8GB | ~12.5M | Windows 11 23H2 |
| GeForce RTX 4070 12GB   | ~9.5M  | Windows 11 23H2 |


| CPU                  |keys/s | OS              |
|:---------------------|:-----:|:----------------|
| Intel Core i7-13700K | ~1M   | Windows 11 23H2 |
| AMD Ryzen 9 7950X    | ~920k | Fedora Linux 42 |
| AMD Ryzen 7 7800x3d  | ~570K | Fedora Linux 42 |
| AMD Ryzen 5 7600X    | ~500K | Debian Linux 12 |
| Apple M1             | ~280k | macOS 26        |
| lx2160a A72          | ~143K | Fedora Linux 42 |