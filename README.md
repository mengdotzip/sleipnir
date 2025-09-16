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
./sleipnir -pattern MyName -ignore-case false

# Verbose logging
./sleipnir -pattern awesome -verbose
```