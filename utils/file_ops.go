package utils

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func ReadFileContent(filepath string) ([]byte, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		log.Errorf("Reading file: `%s` failed with error: %s\n", filepath, err)
		return nil, err
	}
	return content, nil
}

func WriteFileContent(filepath string, fileContent []byte) error {
	err := os.WriteFile(filepath, fileContent, 0600)
	if err != nil {
		return err
	}
	return nil
}

func CheckIfResourcesYAMLExistsAndReturnPath() (bool, string) {
	homeDir, _ := os.UserHomeDir()
	currentDir, _ := os.UserHomeDir()
	if homeDir == "" && currentDir == "" {
		log.Warnf("Error checking the existance of resources.yaml. The app will not able to clean up existing resources.\n")
		return false, ""
	}
	fileExist := false
	filepathToReturn := filepath.Join(homeDir, ".sockv5er", "resources.yaml")
	if homeDir != "" {
		resourcesFilepath := filepath.Join(homeDir, ".sockv5er", "resources.yaml")
		if _, err := os.Stat(resourcesFilepath); errors.Is(err, os.ErrNotExist) {
			fileExist = false
		} else {
			fileExist = true
		}
	} else if fileExist == false && currentDir != "" {
		resourcesFilepath := filepath.Join(currentDir, ".sockv5er", "resources.yaml")
		if _, err := os.Stat(resourcesFilepath); errors.Is(err, os.ErrNotExist) {
			fileExist = false
		} else {
			fileExist = true
			filepathToReturn = resourcesFilepath
		}
	}
	return fileExist, filepathToReturn
}

func CreateSockV5erDirectory() string {
	resourcesDirPath, err := os.UserHomeDir()
	if err != nil {
		log.Error(err)
		log.Warnf("Error getting user's home directory. Storing the resources.yaml file in the current working directory.\n")
		resourcesDirPath, err = os.Getwd()
		if err != nil {
			log.Error(err)
			log.Fatal("Error getting current working directory. Please check the permissions and restart the app.")
		}
	}
	resourcesFilepath := filepath.Join(resourcesDirPath, ".sockv5er")
	err = os.Mkdir(resourcesFilepath, 0750)
	if err != nil && !os.IsExist(err) {
		log.Error(err)
		log.Fatal("Unable to create the directory to store the resources.yaml file.")
	}
	return resourcesFilepath
}
