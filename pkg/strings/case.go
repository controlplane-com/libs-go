package strings

import corestrings "strings"

func Capitalize(str string) string {
	if len(str) < 2 {
		return corestrings.ToUpper(str)
	}
	return corestrings.ToUpper(str[0:1]) + str[1:]
}
