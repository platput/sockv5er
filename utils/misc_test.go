package utils

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"testing"
)

func TestGetRegionFromUserInput(t *testing.T) {
	countryOptions := make([]map[string]string, 0)
	countryOption := make(map[string]string)
	countryOption["region"] = "us-west-1"
	countryOption["country"] = "USA"
	countryOptions = append(countryOptions, countryOption)
	want := "us-west-1"
	got, err := getRegionFromUserInput(countryOptions, 1)
	if err != nil || want != got {
		t.Error("Unexpected output for getRegionFromUserInput.")
	}
}

func TestGetUserInput(t *testing.T) {
	in, err := os.CreateTemp("", "")
	if err != nil {
		log.Fatal(err)
	} else {
		defer func(in *os.File) {
			err := in.Close()
			if err != nil {
				log.Warnf("Testing TestGetUserInput resulted in non fatal errors: %s\n", err)
			}
		}(in)
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				log.Warnf("Testing TestGetUserInput resulted in non fatal errors: %s\n", err)
			}
		}(in.Name())
	}
	_, err = in.WriteString("5")
	if err != nil {
		t.Fatal(err)
	}
	_, err = in.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}
	want := 5
	got := getUserInput(5, in)
	if want != got {
		t.Error("Unexpected output for getUserInput")
	}
}
