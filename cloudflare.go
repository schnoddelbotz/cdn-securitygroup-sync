package main

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func getCloudflareCIDRs() []string {
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://www.cloudflare.com/ips-v4", nil)
	exitIfError("ERROR with cloudflare request", err)
	resp, err := client.Do(req)
	exitIfError("ERROR with cloudflare response", err)
	byt, err := ioutil.ReadAll(resp.Body)
	exitIfError("ERROR reading cloudflare response", err)
	cidrs := strings.Split(string(byt), "\n")
	return cidrs[:len(cidrs)-1]
}
