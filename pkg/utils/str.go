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
	num, err := strconv.ParseUint(str, 10, 64)
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

func RuneIsChinese(r rune) bool {
	if unicode.Is(unicode.Han, r) || (regexp.MustCompile("[\u3002\uff1b\uff0c\uff1a\u201c\u201d\uff08\uff09\u3001\uff1f\u300a\u300b]").MatchString(string(r))) {
		return true
	}
	return false
}

func StrContainsChinese(str string) (exists bool) {
	rs := []rune(str)
	for _, r := range rs {
		if RuneIsChinese(r) {
			exists = true
			return
		}
	}
	return
}

func StrContainsContinuousNum(str string) int {
	arr := strings.Split(regexp.MustCompile("[^0-9]+").ReplaceAllString(str, ""), "")
	l := len(arr)
	maxCount := 0
	if l > 0 {
		count := 0
		lastItem := Str2Int(arr[0])
		direction := 0
		for i := 1; i < l; i++ {
			item := Str2Int(arr[i])
			if item == lastItem+1 {
				if direction == 1 && count >= 0 {
					direction = 0
					count = 0
				}
				count++
			} else if item == lastItem-1 {
				if direction == 0 && count >= 0 {
					direction = 1
					count = 0
				}
				count++
			} else {
				direction = 0
				count = 0
			}
			lastItem = item
			if count+1 > maxCount {
				maxCount = count + 1
			}
		}
	}
	return maxCount
}
