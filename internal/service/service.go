package service

import (
	"fmt"
	"log"
	"strings"
	"url-shortner/internal/models"

	"github.com/yihleego/base62"
)

func MakeShortURL(url string) (string, error) {
	// здесь мы должны закодировать исходный юрл, оба закинуть в мапу
	if (!strings.HasPrefix(url, "https://") || !strings.HasPrefix(url, "http://")) && !strings.Contains(url, ".") {
		log.Printf("URL is incorrect: %s", url)
		return "", fmt.Errorf("URL is incorrect")
	}

	shortURL := encodeURL(url)

	if _, exists := models.URLs[url]; !exists {
		models.URLs[url] = shortURL
	}

	return shortURL, nil
}

func encodeURL(url string) string {
	encoded := base62.StdEncoding.EncodeToString([]byte(url))

	cutEncodedHash := encoded[:5]
	shortURL := fmt.Sprintf("cut.ru/%s", cutEncodedHash)
	return shortURL
}
