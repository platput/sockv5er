package utils

import (
	"errors"
	"github.com/ip2location/ip2location-go/v9"
	"net"
)

type IP2LocationFinder interface {
	GetIP(string) (string, error)
	FindCountry(string) (string, error)
}

type GeoHelper struct {
	Settings *Settings
}

func (h *GeoHelper) GetIP(ep string) (string, error) {
	ips, _ := net.LookupIP(ep)
	epIP := ""
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			epIP = ipv4.String()
		}
	}
	if epIP == "" {
		return "", errors.New("IP could not be found for the EP: " + ep)
	}
	return epIP, nil
}

func (h *GeoHelper) FindCountry(ep string) (string, error) {
	ip, err := h.GetIP(ep)
	if err != nil {
		return "", err
	}
	db, err := ip2location.OpenDB(h.Settings.GeoLocationFile)
	results, err := db.Get_all(ip)
	if err != nil {
		return "", errors.New("Country name couldn't be found for the ep: " + ep)
	}
	return results.Country_long, nil
}
