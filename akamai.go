package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
)

type siteShieldMaps struct {
	SiteShieldMap []siteShieldMap `json:"siteShieldMaps"`
}

type siteShieldMap struct {
	Acknowledged bool     `json:"acknowledged"`
	ID           int      `json:"id"`
	CurrentCidrs []string `json:"currentCidrs"`
}

func getAkamaiConfig() edgegrid.Config {
	return edgegrid.Config{
		Host:         os.Getenv("AKAMAI_EDGEGRID_HOST"),
		ClientToken:  os.Getenv("AKAMAI_EDGEGRID_CLIENT_TOKEN"),
		ClientSecret: os.Getenv("AKAMAI_EDGEGRID_CLIENT_SECRET"),
		AccessToken:  os.Getenv("AKAMAI_EDGEGRID_ACCESS_TOKEN"),
		MaxBody:      1024,
		HeaderToSign: []string{},
		Debug:        false,
	}
}

func getSiteshieldMaps() siteShieldMaps {
	client := http.Client{}
	config := getAkamaiConfig()

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/siteshield/v1/maps", config.Host), nil)
	exitIfError("Akamai request failed", err)
	req = edgegrid.AddRequestHeader(config, req)
	resp, err := client.Do(req)
	exitIfError("Akamai response error", err)
	byt, err := ioutil.ReadAll(resp.Body)
	exitIfError("Unable to read Akamai response", err)

	var ssMaps siteShieldMaps
	err = json.Unmarshal(byt, &ssMaps)
	exitIfError("Decoding JSON failed", err)
	return ssMaps
}

func getSiteshieldMap(ssid int) siteShieldMap {
	ssMaps := getSiteshieldMaps()
	for _, m := range ssMaps.SiteShieldMap {
		if m.ID == ssid {
			return m
		}
	}
	log.Fatalf("Unable to find given siteshield map by ID %d", ssid)
	var voidMap siteShieldMap
	return voidMap
}

func printSSIDs() {
	ssMaps := getSiteshieldMaps()
	for _, m := range ssMaps.SiteShieldMap {
		print(m.ID)
		print("\n")
	}
	os.Exit(0)
}

func acknowledgeCIDRs(ssid int) {
	client := http.Client{}
	config := getAkamaiConfig()

	ackURL := fmt.Sprintf("https://%s/siteshield/v1/maps/%d/acknowledge", config.Host, ssid)
	req, err := http.NewRequest("POST", ackURL, nil)
	exitIfError("Akamai ack request failed", err)
	req = edgegrid.AddRequestHeader(config, req)
	resp, err := client.Do(req)
	exitIfError("Akamai ack response error", err)
	byt, err := ioutil.ReadAll(resp.Body)
	exitIfError("Unable to read Akamai ack response", err)
	log.Printf("ACK response: %s", string(byt))
	// TBD: verify response
}
