# Sleipnir - Lightning-Fast Vanity SSH Key Generator

Sleipnir is a blazing-fast vanity SSH key generator written in Go, capable of generating **470,000+** ED25519 keys per second on modern hardware. Named after Odin's eight-legged horse from Norse mythology, Sleipnir gallops through keyspace at incredible speeds to find your perfect vanity SSH keys.



## Usage

Basic Usage
```bash
# Find "cool" anywhere in the SSH key
./sleipnir -pattern cool

# Find key starting with "AAAA"
./sleipnir -pattern AAAA -location start

# Find key ending with "1337"
./sleipnir -pattern 1337 -location end
```

### Advanced Options

```bash
# Use specific number of workers
./sleipnir -pattern github -workers 16

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
Sleipnir galloping with 12 workers...
Hunting pattern: MENG
Press Ctrl+C to stop
Expected tries: 1.6777216e+07
Average keys per second: 482034 Total tries: 2410173 Calculated wait time: 0d 00h 00m 05s/0d 00h 00m 34s
Average keys per second: 474232 Total tries: 4786627 Calculated wait time: 0d 00h 00m 10s/0d 00h 00m 35s
...
Average keys per second: 481823 Total tries: 16834980 Calculated wait time: 0d 00h 00m 35s/0d 00h 00m 34s
Made it in 18841452 tries

KEY FOUND :)!
OpenSSH Private Key:
-----BEGIN OPENSSH PRIVATE KEY-----
Removed so nobody would actually use this key :p
-----END OPENSSH PRIVATE KEY-----

Public Key:
AAAAC3NzaC1lZDI1NTE5AAAAIK+p9TNjWPHhV55/4LABlUapaCD0jHgPUrsfjdOkMENG
All goroutines closed successfully
```
**NOTE** If you want the PKCS#8  format instead of OpenSSH you will have to run sleipnir with -verbose

## Tests
Benchmark the speed of the ssh keygen per core:
```
go test -bench .
```

Test if we are generating valid ssh keys:
```
go test -v
```
