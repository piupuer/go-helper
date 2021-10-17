package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type IpResp struct {
	Status   string `json:"status"`
	Province string `json:"province"`
	City     string `json:"city"`
}

// get real ip location by amap
func GetIpRealLocation(ip, key string) string {
	resp, err := http.Get(fmt.Sprintf("https://restapi.amap.com/v3/ip?ip=%s&key=%s", ip, key))
	address := "unknown address"
	if err != nil {
		return address
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return address
	}
	var result IpResp
	Json2Struct(string(data), &result)
	if result.Status == "1" {
		address = result.Province
		if result.City != "" && result.Province != result.City {
			address += result.City
		}
	}
	return address
}
