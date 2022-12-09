package utils

import (
	log "github.com/sirupsen/logrus"
)

func GenerateCountryOptions() ([]map[string]string, error) {
	var countryOptions []map[string]string
	e := ENVData{}
	settings, _ := e.Read()
	helper := AWSHelper{}
	err := helper.InitializeAWS(settings)
	if err != nil {
		return countryOptions, err
	}
	regions := helper.GetRegions()
	gh := GeoHelper{Settings: settings}
	for i := range regions {
		country, err := gh.FindCountry(*regions[i].Endpoint)
		if err != nil {
			log.Warnf("Finding country for endpoint %s failed with error %v", *regions[i].Endpoint, err)
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
