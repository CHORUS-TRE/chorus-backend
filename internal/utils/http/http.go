package http

import "strings"

func getSchemeFromURL(url string) string {
	scheme := "http"
	if strings.HasPrefix(url, "https") {
		scheme = "https"
	}
	return scheme
}

func ExtractURLComponents(url string) (string, string, string) {
	scheme := getSchemeFromURL(url)

	url = strings.TrimPrefix(url, scheme+"://")
	url = strings.TrimSuffix(url, scheme+"/")

	parts := strings.Split(url, "/")
	var host, basePath string
	if len(parts) > 0 {
		host = parts[0]
	}

	if len(parts) > 1 {
		basePath = parts[1]
	}
	return host, basePath, scheme
}
