package utils

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Renders an int32 to a string
func I32ToString(i int32) string {
	return strconv.FormatInt(int64(i), 10)
}

// Renders an int64 to a string
func I64ToString(i *int64) string {
	return strconv.FormatInt(*i, 10)
}

func FormatBytes(i *int64) string {
	if *i > 1024000 {
		var bytes int32
		bytes = int32(*i / 1024000)
		return fmt.Sprintf("%s MB", I32ToString(bytes))
	} else if *i > 1024 {
		var bytes int32
		bytes = int32(*i / 1024)
		return fmt.Sprintf("%s KB", I32ToString(bytes))
	} else {
		return fmt.Sprintf("%s B", I64ToString(i))
	}

}

// Renders a string to lower title case (all lower case except for initial chars in each word)
func LowerTitle(s string) string {
	return strings.Title(strings.ToLower(s))
}

// Replaces all regex matches in a string
func ReplaceAllRegex(pattern string, src string, replace string) string {
	return regexp.MustCompile(pattern).ReplaceAllString(src, replace)
}

// Removes all regex matches in a string
func RemoveAllRegex(pattern string, src string) string {
	return ReplaceAllRegex(pattern, src, "")
}

func RemoveAllBeforeLastChar(char string, src string) string {
	lastIndex := strings.LastIndex(src, char)
	if lastIndex > 0 {
		return src[lastIndex+1:]
	}
	return src
}

// Returns the right x chars from a string
func TakeRight(s string, max int) string {
	result := s
	if len(result) > max {
		if max > 1 {
			max -= 1
		}
		start := len(result) - max - 1
		result = "…" + s[start:len(result)]
	}
	return result
}

// Returns the left x chars from a string and hack
func TakeLeft(s string, max int) string {
	result := s
	if len(result) > max {
		if max > 1 {
			max -= 1
		}
		result = s[0:max] + "…"
	}
	return result
}

// Builds a one-line meter using the amount and total values limited to the given width
func BuildAsciiMeterCurrentTotal(portion uint32, total uint32, width int) string {
	const fullChar = "█"
	const emptyChar = "▒"

	full := 0
	if total > 0 {
		ratio := float64(portion) / float64(total)
		ratio = math.Max(0, math.Min(1.0, ratio))
		full = int(math.Round(ratio * float64(width)))
	}

	return strings.Join([]string{
		strings.Repeat(fullChar, full),
		strings.Repeat(emptyChar, width-full),
	}, "")
}
