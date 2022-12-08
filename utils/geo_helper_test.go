package utils

import "testing"

func TestGetCountryShortNameUSA(t *testing.T) {
	gh := GeoHelper{}
	want := "USA"
	got := gh.GetCountryShortName("United States of America")
	if want != got {
		t.Error("Incorrect shortened country name.")
	}
}

func TestGetCountryShortNameUK(t *testing.T) {
	gh := GeoHelper{}
	want := "UK"
	got := gh.GetCountryShortName("United Kingdom of Great Britain and Northern Ireland")
	if want != got {
		t.Error("Incorrect shortened country name.")
	}
}

func TestGetCountryShortNameIndia(t *testing.T) {
	gh := GeoHelper{}
	want := "India"
	got := gh.GetCountryShortName("India")
	if want != got {
		t.Error("Incorrect shortened country name.")
	}
}
