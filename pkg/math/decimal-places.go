package math

import (
	"strconv"
	"strings"
)

func CountLeadingZeroDecimals(f float64) int {
	if f >= 1 {
		return 0
	}
	str := strconv.FormatFloat(f, 'f', -1, 64)
	i := strings.IndexByte(str, '.')
	str = str[i+1:]
	lenStr := len(str)
	leadingZeroes := 0
	for i = 0; i < lenStr; i++ {
		if str[i] != '0' {
			break
		}
		leadingZeroes++
	}
	return leadingZeroes
}
