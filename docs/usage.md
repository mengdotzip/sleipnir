### Advanced Options

```bash
# Find key starting with "Hi". Keep in mind that the starting string "AAAAC3NzaC1lZDI1NTE5AAAAI" is static
# sleipnir will only start searching after that.
./sleipnir -pattern Hi -location start

# Use specific number of workers, by default we use all threads
./sleipnir -pattern github -workers 16

# Continue even when a key is found 
# I strongly suggest also using -output when using stream
./sleipnir -pattern gitarena -stream

# Output the found key to the specified file
./sleipnir -pattern xmr -output test.txt

# Case-sensitive matching
./sleipnir -pattern MyName -ignore-case=false

# Verbose logging + PKCS#8 format private key
./sleipnir -pattern awesome -verbose
```