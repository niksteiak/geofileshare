package main

import (
	"strings"
)

func ParseURLRawQuery(rawQuery string) map[string]string {
	var urlParameters = make(map[string]string)
	paramSets := strings.Split(rawQuery, "&")
	for idx := 0; idx < len(paramSets); idx++ {
		curSet := paramSets[idx]
		paramAttrs := strings.Split(curSet, "=")
		urlParameters[paramAttrs[0]] = paramAttrs[1]
	}

	return urlParameters
}
