package main

import (
	"unicode"
)

func ContainsOnlyLetters(textItem string) bool {
	for _, r := range textItem {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
