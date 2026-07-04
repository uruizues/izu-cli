package allanime

import "strings"

func CleanProviderPath(path string) string {
	// Add clock.json suffix if needed
	if strings.Contains(path, "/clock") && !strings.HasSuffix(path, ".json") {
		path += ".json"
	}
	return path
}

func ExtractProviderURL(decodedPath string) string {
	if strings.HasPrefix(decodedPath, "http") {
		return decodedPath
	}
	return "https://allanime.day" + decodedPath
}
