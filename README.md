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

# Verbose logging
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
Average keys per second: 484188 Total tries: 2420946 Calculated wait time: 0d 00h 00m 05s/0d 00h 00m 34s
...
Average keys per second: 479806 Total tries: 47987448 Calculated wait time: 0d 00h 01m 40s/0d 00h 00m 34s
Made it in 49701882 tries

KEY FOUND :)!
Private Key:
-----BEGIN PRIVATE KEY-----
Removed so nobody would actually use this key :p
-----END PRIVATE KEY-----

Public Key:
AAAAC3NzaC1lZDI1NTE5AAAAIN1ApPRkQ17fpv812nKS5BeDalJ03/V83QNkzS+UMENG
All goroutines closed successfully
```

## Tests
You can run:
```
go test -bench .
```
To benchmark the speed of the ssh keygen per core.