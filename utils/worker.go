package utils

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

func showIntro() {
	fmt.Println(`
                         _           _____
                        | |         | ____|          
	  ___  ___   ___| | ____   _| |__   ___ _ __ 
	 / __|/ _ \ / __| |/ /\ \ / /___ \ / _ \ '__|
	 \__ \ (_) | (__|   <  \ V / ___) |  __/ |   
	 |___/\___/ \___|_|\_\  \_/ |____/ \___|_|
	`)
	fmt.Println("Downloading counties/regions list...")
	fmt.Println("Please wait a moment.")
}

type SocksV5Er struct {
	settings *Settings
	tracker  *ResourceTracker
	repo     CloudProvider
}

func showRegionsOptions(countryOptions []map[string]string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Country", "Region"})
	for i := range countryOptions {
		t.AppendRows([]table.Row{{i + 1, countryOptions[i]["country"], countryOptions[i]["Region"]}})
		t.AppendSeparator()
	}
	t.Render()
}

func getRegionFromUserInput(countryOptions []map[string]string, selection int) (string, error) {
	regionID := selection - 1
	region := countryOptions[regionID]["Region"]
	return region, nil
}

func getUserInput(numberOfRegions int, in *os.File) int {
	if in == nil {
		in = os.Stdin
	}
	regionsRange := numberOfRegions
	var regionID = 0
	fmt.Println("Enter the id of the Region in which you need to create the socks v5 proxy on.")
	fmt.Printf("Default is 1. Range 1-%d: ", numberOfRegions)
	for {
		_, err := fmt.Fscanf(in, "%d\n", &regionID)
		if err != nil {
			log.Fatalf("Unexpected input. %s\n", err)
		} else if regionID > 0 && regionID <= regionsRange {
			break
		} else {
			fmt.Printf("Please choose a number between 1 and %d: ", numberOfRegions)
		}
	}
	return regionID
}

func (s *SocksV5Er) createSocksV5Tunnel() {
	config := SSHConfig{}
	config.PrivateKey = s.repo.GetPrivateKey()
	config.KnownHostsFilepath = s.settings.SSHKnownHostsPath
	config.SSHUsername = s.settings.SSHUserName
	config.SSHHost = s.repo.GetHostIP()
	config.SSHPort = s.settings.SSHPort
	config.SocksV5IP = s.settings.SocksV5Host
	config.SocksV5Port = s.settings.SocksV5Port
	config.StartSocksV5Server()
}

func (s *SocksV5Er) processResourcesTrackerFile(resourcesFilepath string) {
	fmt.Printf("SockV5er resources.yaml file exists at the path `%s`.\n Clean up before continuing?\n"+
		"Y - Recommened Option. N - Not recommended and possibly dangerous.\n"+
		"Y/N? ", resourcesFilepath)
	input := os.Stdin
	cleanFlag := ""
	_, err := fmt.Fscanf(input, "%s\n", &cleanFlag)
	if err != nil {
		log.Fatalf("Unexpected input. %s\n", err)
	}
	err = s.tracker.ReadResourcesFile()
	if err != nil {
		log.Warnf("Reading resources.yaml file failed with error: %s\n", err)
		return
	}
	if strings.ToLower(cleanFlag) == "y" {
		// clean resources from yaml file
		repo := NewAWSProvider()
		resources := s.tracker.GetResources()
		resourcesCount := len(*resources)
		for i := 0; i < resourcesCount; i++ {
			resource := (*resources)[0]
			// Deleting the resource from AWSResources
			repo.PrepareResourcesForDeletion(ToMap(&resource))
			err := repo.DeleteResources(resource.Region, s.settings, s.tracker)
			// Removing the resource from resources.yaml
			if err != nil {
				fmt.Printf(
					"Couldn't delete atleast one resource. Please delete the resources manually.\n. Region: %s\nInstance Id: %s\nKeypair Name: %s\nSecurity Group Id: %s\n",
					resource.Region,
					resource.InstanceId,
					resource.KeyPairId,
					resource.SecurityGroupId,
				)
			}
			err = s.tracker.ReadResourcesFile()
			resources = s.tracker.GetResources()
		}
	} else {
		// proceed without cleaning
		fmt.Printf("Continuing without deleting the resources might incur additional unnecessary charges in your AWSResources account.")
	}
}

func StartWorker() {
	e := ENVData{}
	s := SocksV5Er{}
	s.repo = NewAWSProvider()
	s.settings = e.Read()
	resourcesTrackerFlag, resourcesFilepath := CheckIfResourcesYAMLExistsAndReturnPath()
	s.settings.TrackingFilepath = resourcesFilepath
	s.tracker = GetNewTracker(resourcesFilepath)
	err := s.repo.Initialize(s.settings)
	if resourcesTrackerFlag {
		s.processResourcesTrackerFile(resourcesFilepath)
	} else {
		//	Creating the tracker file
		trackerFilepath := CreateSockV5erDirectory()
		// Create the file
		resourcesFilepath := filepath.Join(trackerFilepath, "resources.yaml")
		file, err := os.Create(resourcesFilepath)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	}
	showIntro()
	countryOptions := s.repo.GetRegions(s.settings)
	showRegionsOptions(countryOptions)
	selection := getUserInput(len(countryOptions), nil)
	region, err := getRegionFromUserInput(countryOptions, selection)
	if err != nil {
		log.Fatalf("SockV5er failed with error: %s\n", err)
	}
	fmt.Printf("Selected Region: %s\n", region)

	if err != nil {
		log.Fatalf("Exiting program, please submit a bug report at: https://github.com/platoputhur/sockv5er for the error: %s", err)
	}
	err = s.repo.CreateResources(region, s.settings, s.tracker)
	if err != nil {
		log.Fatalf("Exiting program, please submit a bug report at: https://github.com/platoputhur/sockv5er for the error: %s", err)
	}
	log.Infoln("Created all the resources required to start the socksv5 server")
	s.createSocksV5Tunnel()
	log.Infoln("Started the socksv5 server. ")
	fmt.Printf("All systems online. You can set up your browser/system to use the socksv5 server.\nDetails: host: %s\nport: %s\n", s.settings.SocksV5Host, s.settings.SocksV5Port)
}
