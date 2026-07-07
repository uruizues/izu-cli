package allanime

import "strings"

// CleanProviderPath ensures the path has the correct suffix for the AllAnime API.
func CleanProviderPath(path string) string {
	// Add clock.json suffix if needed
	if strings.Contains(path, "/clock") && !strings.HasSuffix(path, ".json") {
		path += ".json"
	}
	return path
}

// ExtractProviderURL converts a provider path to a full URL.
// Paths starting with "http" are returned as-is; others are prefixed with the AllAnime base.
func ExtractProviderURL(decodedPath string) string {
	if strings.HasPrefix(decodedPath, "http") {
		return decodedPath
	}
	return "https://allanime.day" + decodedPath
}
