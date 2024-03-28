package goatquery

import (
	"regexp"
	"strings"
)

var filterOperations = map[string]string{
	"eq":       "=",
	"ne":       "<>",
	"contains": "like",
}

func splitString(input string) []string {
	var result []string
	var buffer strings.Builder
	var singleQuote bool

	for i := 0; i < len(input); i++ {
		char := input[i]

		if char == '\'' {
			buffer.WriteRune(rune(char))
			singleQuote = !singleQuote
		} else if !singleQuote && (char == 'a' || char == 'o') && i+1 < len(input) && (input[i-1:i+3] == " and" || input[i-1:i+2] == " or") {
			if buffer.Len() > 0 {
				result = append(result, strings.TrimSpace(buffer.String()))
				buffer.Reset()
			}

			result = append(result, strings.TrimSpace(input[i:i+3]))
			i += 2
		} else {
			buffer.WriteRune(rune(char))
		}
	}

	if buffer.Len() > 0 {
		result = append(result, strings.TrimSpace(buffer.String()))
	}

	return result
}

func splitStringByWhitespace(str string) []string {
	var parts []string
	var sb strings.Builder
	singleQuote := false

	for _, char := range str {
		switch char {
		case ' ':
			if singleQuote {
				sb.WriteRune(char)
			} else {
				parts = append(parts, sb.String())
				sb.Reset()
			}
		case '\'':
			singleQuote = !singleQuote
			sb.WriteRune(char)
		default:
			sb.WriteRune(char)
		}
	}

	parts = append(parts, sb.String())

	return parts
}

func getValueBetweenQuotes(input string) string {
	re := regexp.MustCompile(`'([^']*)'`)
	match := re.FindStringSubmatch(input)
	if len(match) > 1 {
		return match[1]
	}
	return match[0]
}
