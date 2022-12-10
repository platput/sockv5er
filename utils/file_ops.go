package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func ReadFileContent(filepath string) []byte {
	content, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("Reading file: `%s` failed with error: %s", filepath, err)
	}
	return content
}
