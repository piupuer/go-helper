package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetIpRealLocation get real ip location by amap
func GetIpRealLocation(ip, key string) (address string) {
	rp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s?lang=zh-CN", ip))
	address = "unknown address"
	if err != nil {
		return
	}
	defer rp.Body.Close()
	data, err := ioutil.ReadAll(rp.Body)
	if err != nil {
		return
	}
	var result map[string]interface{}
	Json2Struct(string(data), &result)
	if result["status"] == "success" {
		country := result["country"].(string)
		city := result["city"].(string)
		address = country + city
	}
	return
}
