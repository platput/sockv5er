package main

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/platoputhur/sockv5er/utils"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	countryOptions := showIntro()
	fmt.Println("Enter the id of the region in which you need to create the socks v5 proxy on.")
	fmt.Print("Default is 14. Range 1-16: ")
	var regionID int
	_, err := fmt.Scanln(&regionID)
	regionID = regionID - 1
	if err != nil {
		log.Fatalf("Unexpected error, please restart the application. Error: %s\n", err)
	}
	fmt.Printf("Selected region: %s\nCountry: %s\n", countryOptions[regionID]["region"], countryOptions[regionID]["country"])
}

func showIntro() []map[string]string {
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
	countryOptions, err := utils.GenerateCountyOptions()
	if err != nil {
		log.Fatalf("Generating country options failed with error: %s\n", err)
	}
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Country", "Region"})
	for i := range countryOptions {
		t.AppendRows([]table.Row{{i + 1, countryOptions[i]["country"], countryOptions[i]["region"]}})
		t.AppendSeparator()
	}
	t.Render()
	return countryOptions
}
