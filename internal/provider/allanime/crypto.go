package allanime

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const EncryptKey = "Xot36i3lK3:v1"

func deriveKey() []byte {
	hash := sha256.Sum256([]byte(EncryptKey))
	return hash[:]
}

func decryptTobeparsed(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	if len(data) < 17 {
		return "", fmt.Errorf("data too short")
	}

	// Skip first byte (version marker)
	_ = data[0]

	// IV = bytes 1-12 + [0,0,0,2]
	iv := make([]byte, 16)
	copy(iv, data[1:13])
	iv[12] = 0
	iv[13] = 0
	iv[14] = 0
	iv[15] = 2

	// Remove last 16 bytes (auth tag) and first 13 bytes (version + IV)
	ciphertext := data[13 : len(data)-16]

	key := deriveKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes cipher failed: %w", err)
	}

	stream := cipher.NewCTR(block, iv)

	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return string(plaintext), nil
}

func decodeHexMapping(hexStr string) (string, error) {
	hexMap := map[string]string{
		"79": "A", "7a": "B", "7b": "C", "7c": "D", "7d": "E", "7e": "F", "7f": "G",
		"70": "H", "71": "I", "72": "J", "73": "K", "74": "L", "75": "M", "76": "N",
		"77": "O", "68": "P", "69": "Q", "6a": "R", "6b": "S", "6c": "T", "6d": "U",
		"6e": "V", "6f": "W", "60": "X", "61": "Y", "62": "Z",
		"59": "a", "5a": "b", "5b": "c", "5c": "d", "5d": "e", "5e": "f", "5f": "g",
		"50": "h", "51": "i", "52": "j", "53": "k", "54": "l", "55": "m", "56": "n",
		"57": "o", "48": "p", "49": "q", "4a": "r", "4b": "s", "4c": "t", "4d": "u",
		"4e": "v", "4f": "w", "40": "x", "41": "y", "42": "z",
		"08": "0", "09": "1", "0a": "2", "0b": "3", "0c": "4",
		"0d": "5", "0e": "6", "0f": "7", "00": "8", "01": "9",
		"15": "-", "16": ".", "67": "_", "46": "~", "02": ":", "17": "/",
	}

	// Strip prefix if present
	if len(hexStr) > 2 && hexStr[:2] == "--" {
		hexStr = hexStr[2:]
	}

	result := ""
	for i := 0; i < len(hexStr)-1; i += 2 {
		pair := hexStr[i : i+2]
		if val, ok := hexMap[pair]; ok {
			result += val
		} else {
			// Fallback: try standard hex decode
			b, err := hex.DecodeString(pair)
			if err == nil && len(b) > 0 {
				result += string(b)
			}
		}
	}

	return result, nil
}
