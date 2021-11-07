package utils

import (
	"encoding/base64"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	camelRe = regexp.MustCompile("(_)([a-zA-Z]+)")
	snakeRe = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func CamelCase(str string) string {
	camel := camelRe.ReplaceAllString(str, " $2")
	camel = strings.Title(camel)
	camel = strings.Replace(camel, " ", "", -1)
	return camel
}

func CamelCaseLowerFirst(str string) string {
	camel := CamelCase(str)
	for i, v := range camel {
		return string(unicode.ToLower(v)) + camel[i+1:]
	}
	return camel
}

func SnakeCase(str string) string {
	snake := snakeRe.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
}

func RemoveRepeat(arr []string) []string {
	newArr := make([]string, 0, len(arr))
	temp := map[string]struct{}{}
	for _, item := range arr {
		if _, ok := temp[item]; !ok {
			// struct{}{} no memory usage
			temp[item] = struct{}{}
			newArr = append(newArr, item)
		}
	}
	return newArr
}

func Str2Uint(str string) uint {
	num, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0
	}
	return uint(num)
}

func Str2UintArr(str string) []uint {
	ids := make([]uint, 0)
	s := strings.TrimSpace(str)
	if s == "" {
		return ids
	}
	idArr := strings.Split(s, ",")
	for _, v := range idArr {
		ids = append(ids, Str2Uint(v))
	}
	return ids
}

func Str2Int(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return num
}

func Str2Int64(str string) int64 {
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return num
}

func Str2IntArr(str string) []int {
	ids := make([]int, 0)
	s := strings.TrimSpace(str)
	if s == "" {
		return ids
	}
	idArr := strings.Split(s, ",")
	for _, v := range idArr {
		ids = append(ids, Str2Int(v))
	}
	return ids
}

func Str2Int64Arr(str string) []int64 {
	ids := make([]int64, 0)
	s := strings.TrimSpace(str)
	if s == "" {
		return ids
	}
	idArr := strings.Split(s, ",")
	for _, v := range idArr {
		ids = append(ids, Str2Int64(v))
	}
	return ids
}

func EncodeStr2Base64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func DecodeStrFromBase64(str string) string {
	decodeBytes, _ := base64.StdEncoding.DecodeString(str)
	return string(decodeBytes)
}
