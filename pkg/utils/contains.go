package utils

import "github.com/thoas/go-funk"

// Contains whether the array contains interface
func Contains(arr interface{}, item interface{}) bool {
	switch arr.(type) {
	case []uint:
		// funk not implement ContainsUint
		if val, ok := item.(uint); ok {
			return ContainsUint(arr.([]uint), val)
		}
		break
	case []int:
		if val, ok := item.(int); ok {
			return funk.ContainsInt(arr.([]int), val)
		}
		break
	case []string:
		if val, ok := item.(string); ok {
			return funk.ContainsString(arr.([]string), val)
		}
		break
	case []int32:
		if val, ok := item.(int32); ok {
			return funk.ContainsInt32(arr.([]int32), val)
		}
		break
	case []int64:
		if val, ok := item.(int64); ok {
			return funk.ContainsInt64(arr.([]int64), val)
		}
		break
	case []float32:
		if val, ok := item.(float32); ok {
			return funk.ContainsFloat32(arr.([]float32), val)
		}
		break
	case []float64:
		if val, ok := item.(float64); ok {
			return funk.ContainsFloat64(arr.([]float64), val)
		}
		break
	}
	// funk use reflect as default, performance is not as good as type asserts
	return funk.Contains(arr, item)
}

// ContainsUint whether the array contains uint
func ContainsUint(arr []uint, item uint) bool {
	for _, v := range arr {
		if v == item {
			return true
		}
	}
	return false
}

// ContainsUintIndex whether the array contains uint return index or return -1
func ContainsUintIndex(arr []uint, item uint) int {
	for i, v := range arr {
		if v == item {
			return i
		}
	}
	return -1
}

// ContainsUintThenRemove whether the array contains uint and remove it
func ContainsUintThenRemove(arr []uint, item uint) []uint {
	index := ContainsUintIndex(arr, item)
	if index >= 0 {
		arr = append(arr[:index], arr[index+1:]...)
	}
	return arr
}
