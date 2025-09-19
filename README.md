# Sleipnir - Lightning-Fast Vanity SSH Key Generator

Sleipnir is a blazing-fast vanity SSH key generator written in Go, capable of generating **1,000,000+** ED25519 keys per second on modern hardware. Named after Odin's eight-legged horse from Norse mythology, Sleipnir gallops through keyspace at incredible speeds to find your perfect vanity SSH keys.



## Usage

Basic Usage
```bash
# Find "cool" anywhere in the SSH key
./sleipnir -pattern cool

# Find key starting with "Hi". Keep in mind that the starting string "AAAAC3NzaC1lZDI1NTE5AAAAI" is static
# sleipnir will only start searching after that.
./sleipnir -pattern Hi -location start

# Find key ending with "1337" OR "meng" OR "github"
./sleipnir -pattern 1337,meng,github -location end
```

### Advanced Options

```bash
# Use specific number of workers, by default we use all threads
./sleipnir -pattern github -workers 16

# Continue even when a key is found 
# I strongly suggest also using -output when using stream
./sleipnir -pattern mari -stream

# Output the found key to the specified file
./sleipnir -pattern xmr -output test.txt

# Case-sensitive matching
./sleipnir -pattern MyName -ignore-case=false

# Verbose logging + PKCS#8 format private key
./sleipnir -pattern awesome -verbose
```

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

| CPU                  |keys/s | OS              |
|:---------------------|:-----:|:----------------|
| Intel Core i7-13700K | ~1M   | Windows 11 23H2 |
| AMD Ryzen 9 7950X    | ~920k | Fedora Linux 42 |
| AMD Ryzen 7 7800x3d  | ~570K | Fedora Linux 42 |
| AMD Ryzen 5 7600X    | ~500K | Debian Linux 12 |
| Apple M1             | ~280k | macOS 26        |
| lx2160a A72          | ~143K | Fedora Linux 42 |
