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

func unescapeTextContent(content string, allowedToken tokenType, hasNextToken bool) string {
	// Unescape `#, `${ and `$(
	if allowedToken >= envVarToken {
		content = strings.ReplaceAll(content, "`"+envVarPrefix, envVarPrefix)
	}

	if allowedToken >= executableToken {
		content = strings.ReplaceAll(content, "`"+executablePrefix, executablePrefix)
	}

	content = strings.ReplaceAll(content, "`"+commentPrefix, commentPrefix)

	// Handle escaped backtick at the end
	if hasNextToken && strings.HasSuffix(content, "\\`") {
		content = strings.TrimSuffix(content, "\\`") + "`"
	}

	return content
}

func isStartOfToken(tokenTypePrefix, prev, rest string) bool {
	return strings.HasPrefix(rest, tokenTypePrefix) && (!strings.HasSuffix(prev, "`") || strings.HasSuffix(prev, "\\`"))
}

func Tokenize(input string, allowedToken tokenType) ([]token, string) {
	result := []token{}
	inputRunes := []rune(input)

	currentContent := ""
	var currentTokenType tokenType = textToken
	var executableQuoteRune rune
	var executableQuoteEnd int

	idx := 0
	for idx < len(inputRunes) {
		rest := string(inputRunes[idx:])
		prev := string(inputRunes[:idx])

		if currentTokenType == textToken {
			// EnvVar
			if allowedToken >= envVarToken && isStartOfToken(envVarPrefix, prev, rest) {
				idx += len(envVarPrefix)

				currentTokenType = envVarToken
			}

			// Executable
			if allowedToken >= executableToken && isStartOfToken(executablePrefix, prev, rest) {
				idx += len(executablePrefix)

				currentTokenType = executableToken
			}

			// Comment
			if isStartOfToken(commentPrefix, prev, rest) {
				idx += len(commentPrefix)

				currentTokenType = commentToken
			}

			if currentTokenType != textToken {
				if len(currentContent) > 0 {
					// Pack up collected text
					result = append(result, token{
						tokenType:    textToken,
						content:      unescapeTextContent(currentContent, allowedToken, true),
						fatalContent: currentContent,
					})
				}

				// Comment applies to the rest of the line
				if currentTokenType == commentToken {
					result = append(result, token{
						tokenType:    commentToken,
						fatalContent: commentPrefix + string(inputRunes[idx:]),
					})

					return result, ""
				}

				currentContent = ""
				continue
			}
		}

		if currentTokenType == envVarToken && isStartOfToken("}", prev, rest) {
			unescapedContent := strings.ReplaceAll(currentContent, "`}", "}")

			if strings.HasSuffix(unescapedContent, "\\`") {
				unescapedContent = strings.TrimSuffix(unescapedContent, "\\`") + "`"
			}

			result = append(result, token{
				tokenType:    envVarToken,
				content:      unescapedContent,
				fatalContent: envVarPrefix + currentContent + "}",
			})

			currentTokenType = textToken
			currentContent = ""

			idx += 1
			continue
		}

		if currentTokenType == executableToken {
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
					fatalContent: executablePrefix + currentContent + ")",
				})

				currentTokenType = textToken
				currentContent = ""

				idx += 1
				continue
			}
		}

		currentContent += rest[:1]
		idx += 1
	}

	if currentTokenType == envVarToken {
		result = append(result, token{
			tokenType:    errorToken,
			fatalContent: envVarPrefix + currentContent,
		})

		return result, fmt.Sprintf("Missing closing bracket for environment variable: %s%s", envVarPrefix, currentContent)
	}

	if currentTokenType == executableToken {
		result = append(result, token{
			tokenType:    errorToken,
			fatalContent: executablePrefix + currentContent,
		})

		if executableQuoteRune != 0 {
			return result, fmt.Sprintf("Unterminated quote sequence for executable: %s%s", executablePrefix, currentContent)
		}

		return result, fmt.Sprintf("Missing closing parenthesis for executable: %s%s", executablePrefix, currentContent)
	}

	if len(currentContent) > 0 && currentTokenType == textToken {
		return append(result, token{
			tokenType:    textToken,
			content:      unescapeTextContent(currentContent, allowedToken, false),
			fatalContent: currentContent,
		}), ""
	}

	return result, ""
}
