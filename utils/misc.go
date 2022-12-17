package utils

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"os"
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
	fmt.Println("Downloading counties/regions list from AWS...")
	fmt.Println("Please wait a moment.")
}

type SocksV5Er struct {
	settings  *Settings
	awsHelper *AWSHelper
}

func (s *SocksV5Er) generateCountryOptions() ([]map[string]string, error) {
	var countryOptions []map[string]string
	regions := s.awsHelper.GetRegions()
	gh := GeoHelper{Settings: s.settings}
	for i := range regions {
		country, err := gh.FindCountry(*regions[i].Endpoint)
		if err != nil {
			log.Warnf("Finding country for endpoint %s failed with error %v\n", *regions[i].Endpoint, err)
		} else {
			country := gh.GetCountryShortName(country)
			countryToRegionMap := map[string]string{
				"country": country,
				"region":  *regions[i].RegionName,
			}
			countryOptions = append(countryOptions, countryToRegionMap)
		}
	}
	return countryOptions, nil
}

func (s *SocksV5Er) getRegionsAndCountries() []map[string]string {
	countryOptions, err := s.generateCountryOptions()
	if err != nil {
		log.Fatalf("Generating country options failed with error: %s\n", err)
	}
	if len(countryOptions) < 1 {
		log.Fatalf("Generating country options failed please try again after some time.\n")
	}
	return countryOptions
}

func showRegionsOptions(countryOptions []map[string]string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Country", "Region"})
	for i := range countryOptions {
		t.AppendRows([]table.Row{{i + 1, countryOptions[i]["country"], countryOptions[i]["region"]}})
		t.AppendSeparator()
	}
	t.Render()
}

func getRegionFromUserInput(countryOptions []map[string]string, selection int) (string, error) {
	regionID := selection - 1
	region := countryOptions[regionID]["region"]
	return region, nil
}

func getUserInput(numberOfRegions int, in *os.File) int {
	if in == nil {
		in = os.Stdin
	}
	regionsRange := numberOfRegions
	var regionID = 0
	fmt.Println("Enter the id of the region in which you need to create the socks v5 proxy on.")
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

func createSocksV5Tunnel() {
	config := SSHConfig{}
	e := ENVData{}
	settings, _ := e.Read()
	config.SocksV5IP = "127.0.0.1"
	config.SocksV5Port = "1337"
	config.SSHHost = "techtuft.com"
	config.SSHPort = "22"
	privateKey, err := ReadFileContent(settings.PrivateKeyPath)
	if err != nil {
		log.Fatal(err)
	}
	config.PrivateKey = privateKey
	config.KnownHostsFilepath = settings.SSHKnownHostsPath
	config.SSHUsername = "ubuntu"
	config.StartSocksV5Server()
}

func processResourcesYAMLFileIfExists() {
	status, resourcesFilepath := CheckIfResourcesYAMLExistsAndReturnPath()
	if status == true {
		fmt.Printf("SockV5er resources.yaml file exists at the path `%s`.\n Clean up before continuing?\n"+
			"Y - Recommened Option. N - Not recommended and possibly dangerous.\n"+
			"Y/N? ", resourcesFilepath)
		input := os.Stdin
		cleanFlag := ""
		_, err := fmt.Fscanf(input, "%d\n", &cleanFlag)
		if err != nil {
			log.Fatalf("Unexpected input. %s\n", err)
		}
		var tracker ResourcesTracker
		tracker = &YAMLHelper{
			filepath: resourcesFilepath,
		}
		err = tracker.ReadResourcesFile()
		if err != nil {
			return
		}
		cloudHelper := CloudHelper{awsHelper: &awsHelper}
		if strings.ToLower(cleanFlag) == "y" {
			// clean resources from yaml file
			resources := *tracker.GetResources()
			for i := range resources {
				resource := &resources[i]
				err = cloudHelper.DeleteResource(resource)
				tracker.RemoveResource(resource)
				if err != nil {
					log.Warnf(
						"Couldn't delete atleast one resource. Please delete the resources manually.\n. Region: %s\nInstance Id: %s\nKeypair Name: %s\nSecurity Group Id: %s\n",
						resource.Region,
						resource.InstanceId,
						resource.KeyPairName,
						resource.SecurityGroupId,
					)
				}
			}
		} else {
			// proceed without cleaning
			log.Warnln("Continuing without deleting the resources might incur additional unnecessary charges in your AWS account.")
		}
	}
}

var awsHelper = AWSHelper{}

func StartWorker() {
	s := SocksV5Er{}
	s.awsHelper = &awsHelper
	e := ENVData{}
	s.settings, _ = e.Read()
	err := awsHelper.InitializeAWS(s.settings)
	processResourcesYAMLFileIfExists()
	showIntro()
	countryOptions := s.getRegionsAndCountries()
	showRegionsOptions(countryOptions)
	selection := getUserInput(len(countryOptions), nil)
	region, err := getRegionFromUserInput(countryOptions, selection)
	if err != nil {
		log.Fatalf("SockV5er failed with error: %s\n", err)
	}
	fmt.Printf("Selected region: %s\n", region)
	createSocksV5Tunnel()
}
