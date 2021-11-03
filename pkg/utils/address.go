package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

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
	var result map[string]string
	Json2Struct(string(data), &result)
	if result["status"] == "1" {
		address = result["province"]
		if result["city"] != "" && address != result["city"] {
			address += result["province"]
		}
	}
	return address
}
