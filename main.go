package main

import (
	"fmt"
	"github.com/platoputhur/sockv5er/utils"
)

func main() {
	e := utils.ENVData{}
	settings, _ := e.Read()
	helper := utils.AWSHelper{}
	err := helper.InitializeAWS(settings)
	if err != nil {
		return
	}

	regions := helper.GetRegions()
	gh := utils.GeoHelper{Settings: settings}
	for i := range regions {
		country, err := gh.FindCountry(*regions[i].Endpoint)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%s - %s\n", country, *regions[i].RegionName)
		}
	}
}
