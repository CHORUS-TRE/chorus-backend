package utils

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

func FromString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ToStrings(inputStrings []*string) []string {
	outputStrings := make([]string, len(inputStrings))
	for i, value := range inputStrings {
		outputStrings[i] = ToString(value)
	}
	return outputStrings
}

func UintsToStrings(inputStrings []uint64) []string {
	outputStrings := make([]string, len(inputStrings))
	for i, value := range inputStrings {
		outputStrings[i] = strconv.FormatUint(value, 10)
	}
	return outputStrings
}

func IntsToStrings(inputStrings []int64) []string {
	outputStrings := make([]string, len(inputStrings))
	for i, value := range inputStrings {
		outputStrings[i] = strconv.FormatInt(value, 10)
	}
	return outputStrings
}

func BoolsToStrings(inputStrings []bool) []string {
	outputStrings := make([]string, len(inputStrings))
	for i, value := range inputStrings {
		outputStrings[i] = strconv.FormatBool(value)
	}
	return outputStrings
}

func StringsToUints(inputStrings []string) ([]uint64, error) {
	output := make([]uint64, len(inputStrings))
	for i, value := range inputStrings {
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, err
		}
		output[i] = uintVal
	}
	return output, nil
}

func ToTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func FromTime(t time.Time) *time.Time {
	return &t
}

func FromNullableTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func ToUint64(u *uint64) uint64 {
	if u == nil {
		return 0
	}
	return *u
}

func ToUint64s(us []*uint64) []uint64 {
	var arr []uint64
	for _, u := range us {
		arr = append(arr, *u)
	}
	return arr
}

func FromBool(b bool) *bool {
	return &b
}

func ToBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func FromUint64(u uint64) *uint64 {
	if u == 0 {
		return nil
	}
	return &u
}

func ConvertStringToUint64(value string) (uint64, error) {
	if value == "" {
		return 0, nil
	}
	n, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func ConvertStringToInt64(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func MustConvertStringToUint64(value string) uint64 {
	n, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func MustConvertStringToTime(layout string, value string) time.Time {
	t, err := time.Parse(layout, value)
	var timeValue time.Time
	if err != nil {
		return timeValue
	}
	return t
}

func ConvertUint64ToString(value uint64) string {
	stringValue := ""
	if value > 0 {
		stringValue = strconv.FormatUint(value, 10)
	}
	return stringValue
}

func StringsToLower(values []string) []string {
	if values == nil {
		return nil
	}
	lowerStrings := make([]string, len(values))
	for i, v := range values {
		lowerStrings[i] = strings.ToLower(v)
	}
	return lowerStrings
}

func ToJsonString(i interface{}) string {
	j, err := json.Marshal(i)
	if err != nil {
		return ""
	}

	return string(j)
}
