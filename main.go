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
	for i := range regions {
		fmt.Println(*regions[i].RegionName)
	}
}
