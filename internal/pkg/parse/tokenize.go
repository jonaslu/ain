package parse

import (
	"fmt"
	"strings"
)

type tokenType int

const (
	errorToken      = 0
	commentToken    = 1
	textToken       = 2
	executableToken = 3
	envVarToken     = 4
)

type token struct {
	tokenType tokenType
	content   string
	// Used in formatting fatals - contains the
	// original untokenized line (for keeping escaped
	// tokens which we loose when removing the escaping).
	fatalContent string
}

const (
	commentPrefix    = "#"
	envVarPrefix     = "${"
	executablePrefix = "$("
)

func isStartOfToken(tokenTypePrefix, prev, rest string) bool {
	return strings.HasPrefix(rest, tokenTypePrefix) && (!strings.HasSuffix(prev, "`") || strings.HasSuffix(prev, "\\`"))
}

func splitTextOnComment(input string) (string, string) {
	inputRunes := []rune(input)

	currentContent := ""
	idx := 0

	for idx < len(inputRunes) {
		rest := string(inputRunes[idx:])
		prev := string(inputRunes[:idx])

		if isStartOfToken(commentPrefix, prev, rest) {
			return currentContent, rest
		}

		currentContent += string(inputRunes[idx])
		idx++
	}

	return currentContent, ""
}

func unescapeEnvVars(content string, hasNextToken bool) string {
	content = strings.ReplaceAll(content, "`"+envVarPrefix, envVarPrefix)

	// Handle escaped backtick at the end
	if hasNextToken && strings.HasSuffix(content, "\\`") {
		content = strings.TrimSuffix(content, "\\`") + "`"
	}

	return content
}

// tokenizeEnvVars does not handle comments, input
// is the content of an expandedSectionLine
func tokenizeEnvVars(input string) ([]token, string) {
	result := []token{}
	inputRunes := []rune(input)

	currentContent := ""
	isEnvVar := false
	idx := 0

	for idx < len(inputRunes) {
		rest := string(inputRunes[idx:])
		prev := string(inputRunes[:idx])

		if !isEnvVar && isStartOfToken(envVarPrefix, prev, rest) {
			if len(currentContent) > 0 {
				result = append(result, token{
					tokenType:    textToken,
					content:      unescapeEnvVars(currentContent, true),
					fatalContent: currentContent,
				})

				currentContent = ""
			}

			idx += len(envVarPrefix)
			isEnvVar = true
			continue
		}

		if isEnvVar && isStartOfToken("}", prev, rest) {
			unescapedContent := strings.ReplaceAll(currentContent, "`}", "}")

			if strings.HasSuffix(unescapedContent, "\\`") {
				unescapedContent = strings.TrimSuffix(unescapedContent, "\\`") + "`"
			}

			result = append(result, token{
				tokenType:    envVarToken,
				content:      unescapedContent,
				fatalContent: envVarPrefix + currentContent + "}",
			})

			isEnvVar = false
			currentContent = ""

			idx += 1
			continue
		}

		currentContent += string(inputRunes[idx : idx+1])
		idx += 1
	}

	if isEnvVar {
		return nil, fmt.Sprintf("Missing closing bracket for environment variable: %s%s", envVarPrefix, currentContent)
	}

	if len(currentContent) > 0 {
		result = append(result, token{
			tokenType:    textToken,
			content:      unescapeEnvVars(currentContent, false),
			fatalContent: currentContent,
		})
	}

	return result, ""
}

func unescapeExecutables(content string, hasNextToken bool) string {
	content = strings.ReplaceAll(content, "`"+executablePrefix, executablePrefix)

	if hasNextToken && strings.HasSuffix(content, "\\`") {
		content = strings.TrimSuffix(content, "\\`") + "`"
	}

	return content
}

func tokenizeExecutables(input string) ([]token, string) {
	result := []token{}
	inputRunes := []rune(input)

	var executableQuoteRune rune
	var executableQuoteEnd int
	executableStartIdx := -1

	currentContent := ""
	idx := 0

	for idx < len(inputRunes) {
		rest := string(inputRunes[idx:])
		prev := string(inputRunes[:idx])

		if executableStartIdx == -1 && isStartOfToken(executablePrefix, prev, rest) {
			if len(currentContent) > 0 {
				result = append(result, token{
					tokenType:    textToken,
					content:      unescapeExecutables(currentContent, true),
					fatalContent: currentContent,
				})

				currentContent = ""
			}

			executableStartIdx = idx

			idx += len(envVarPrefix)
			continue
		}

		if executableStartIdx >= 0 {
			nextRune := []rune(rest)[0]
			switch nextRune {
			case '"', '\'':
				if executableQuoteRune == 0 {
					executableQuoteRune = nextRune

					unescapedContentTillNow := currentContent[executableQuoteEnd:]
					currentContent = currentContent[:executableQuoteEnd] + strings.ReplaceAll(unescapedContentTillNow, "`)", ")")
				} else if !strings.HasSuffix(prev, `\`) && executableQuoteRune == nextRune {
					executableQuoteRune = 0
					executableQuoteEnd = len(currentContent) - 1
				}
			}

			if executableQuoteRune == 0 && isStartOfToken(")", prev, rest) {
				unescapedContentTillNow := currentContent[executableQuoteEnd:]
				currentContent = currentContent[:executableQuoteEnd] + strings.ReplaceAll(unescapedContentTillNow, "`)", ")")
				executableQuoteEnd = 0

				if strings.HasSuffix(currentContent, "\\`") {
					currentContent = strings.TrimSuffix(currentContent, "\\`") + "`"
				}

				result = append(result, token{
					tokenType:    executableToken,
					content:      currentContent,
					fatalContent: string(inputRunes[executableStartIdx : idx+1]),
				})

				executableStartIdx = -1
				currentContent = ""

				idx += 1
				continue
			}
		}

		currentContent += string(inputRunes[idx : idx+1])
		idx += 1
	}

	if executableStartIdx >= 0 {
		if executableQuoteRune != 0 {
			return nil, fmt.Sprintf("Unterminated quote sequence for executable: %s", string(inputRunes[executableStartIdx:]))
		}
		return nil, fmt.Sprintf("Missing closing parenthesis for executable: %s", string(inputRunes[executableStartIdx:]))
	}

	if len(currentContent) > 0 {
		result = append(result, token{
			tokenType:    textToken,
			content:      unescapeExecutables(currentContent, false),
			fatalContent: currentContent,
		})
	}

	return result, ""
}
