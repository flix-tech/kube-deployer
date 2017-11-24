package main

import (
	"github.com/gosimple/slug"
	"regexp"
	"strings"
)

const DNS_MAX_LENGTH = 63

func MakeUrlSlug(name string, length int) string {
	sanitizedName := strings.ToLower(name)

	// Remove non a-z chars in the beginning
	charIndex := 0
	for i := 0; i < len(sanitizedName); i++ {
		if isAZ(rune(sanitizedName[i])) {
			break
		}

		charIndex = i
	}
	sanitizedName = sanitizedName[charIndex:]

	// Remove all non valid dns chars
	sanitizedNameFiltered := ""
	for i := 0; i < len(sanitizedName); i++ {
		char := rune(sanitizedName[i])

		if isDnsValid(char) {
			sanitizedNameFiltered += string(char)
		}
	}
	sanitizedName = sanitizedNameFiltered

	slug.MaxLength = length

	return slug.Make(sanitizedName)
}

func isDnsValid(char rune) bool {
	return isAZ09(char) || char == '-'
}

func isAZ09(char rune) bool {
	return isAZ(char) || is09(char)
}

func isAZ(char rune) bool {
	return regexp.MustCompile(`^[A-Za-z]`).MatchString(string(char))
}

func is09(char rune) bool {
	return regexp.MustCompile(`^[0-9]`).MatchString(string(char))
}
