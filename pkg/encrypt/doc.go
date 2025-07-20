/*
Package encrypt provides AES-GCM encryption utilities with key prefix tagging for different key lengths.

# Supported Key Sizes

AES128 (16 bytes)

AES192 (24 bytes)

AES256 (32 bytes)

# Key Functions

Encrypt(plaintext, key) - Encrypts text and prepends an AES prefix (e.g. `aes256:`).

Decrypt(ciphertext, key) - Verifies prefix, decrypts ciphertext.

GenerateRandomAESKey(bits) - Generates a random printable AES key of the desired size.

ValidateAESKey(key) - Validates key length.

# Prefixes

All encrypted strings are prefixed with:

aes128:

aes192:

aes256:
*/
package encrypt
