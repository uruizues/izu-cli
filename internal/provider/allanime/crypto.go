package allanime

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
)

const EncryptKey = "Xot36i3lK3:v1"

func deriveKey() []byte {
	hash := sha256.Sum256([]byte(EncryptKey))
	return hash[:]
}

// decryptTobeparsed decodes a base64-encoded, AES-GCM encrypted tobeparsed field.
// Matches anipy-cli's _decode_tobeparsed: key=sha256(EncryptKey), nonce=raw[1:13], ciphertext=raw[13:-16], tag=raw[-16:]
func decryptTobeparsed(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	if len(data) < 29 { // 1 version + 12 nonce + 1 ciphertext + 16 tag minimum
		return "", fmt.Errorf("data too short for AES-GCM")
	}

	// Skip first byte (version marker), nonce = bytes 1-12
	nonce := data[1:13]
	// Ciphertext = bytes 13 to len-16, tag = last 16 bytes
	ciphertext := data[13 : len(data)-16]
	tag := data[len(data)-16:]

	key := deriveKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes cipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm failed: %w", err)
	}

	// GCM expects ciphertext || tag
	plaintext, err := gcm.Open(nil, nonce, append(ciphertext, tag...), nil)
	if err != nil {
		return "", fmt.Errorf("gcm decrypt failed: %w", err)
	}

	return string(plaintext), nil
}

// decryptSourceURL decodes an AllAnime source URL using the XOR cipher.
// Each pair of hex chars is converted to an integer, XORed with 56,
// then the result is converted to an octal string and interpreted as a character code.
// This matches anipy-cli's _decrypt static method.
func decryptSourceURL(hexStr string) (string, error) {
	// Strip "--" prefix if present
	if len(hexStr) > 2 && hexStr[:2] == "--" {
		hexStr = hexStr[2:]
	}

	if len(hexStr)%2 != 0 {
		return "", fmt.Errorf("hex string has odd length: %d", len(hexStr))
	}

	result := make([]byte, 0, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		pair := hexStr[i : i+2]
		dec, err := strconv.ParseInt(pair, 16, 64)
		if err != nil {
			return "", fmt.Errorf("invalid hex pair %q: %w", pair, err)
		}

		xor := int(dec) ^ 56
		// Convert to 3-digit octal, then interpret as octal number
		octStr := fmt.Sprintf("%03o", xor)
		charCode, err := strconv.ParseInt(octStr, 8, 64)
		if err != nil {
			return "", fmt.Errorf("octal conversion failed: %w", err)
		}

		result = append(result, byte(charCode))
	}

	return string(result), nil
}

// ParseEpisodeResponse parses the raw JSON response, handling tobeparsed decryption if present.
func ParseEpisodeResponse(data []byte) (map[string]interface{}, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	// Check for tobeparsed field at data level
	if dataMap, ok := raw["data"].(map[string]interface{}); ok {
		if tb, ok := dataMap["tobeparsed"].(string); ok && tb != "" {
			decrypted, err := decryptTobeparsed(tb)
			if err != nil {
				return nil, fmt.Errorf("decrypt tobeparsed: %w", err)
			}
			var decryptedData interface{}
			if err := json.Unmarshal([]byte(decrypted), &decryptedData); err != nil {
				return nil, fmt.Errorf("unmarshal decrypted: %w", err)
			}
			raw["data"] = decryptedData
		}
	}

	return raw, nil
}
