package utils

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
)

const quoteEscapeRune = '\\'

var quoteRunes = [...]rune{'"', '\''}

// TODO Do I need the removeQuotes or should I revert it?
func TokenizeLine(commandLine string, removeQuotes bool) ([]string, error) {
	var tokenizedLines []string
	commandLineBytes := []byte(commandLine)

	startWordMarker := -1

	var width int
	var lastRune rune
	var lastQuoteRune rune

	for head := 0; head < len(commandLine); head += width {
		var headRune rune
		headRune, width = utf8.DecodeRune(commandLineBytes[head:])

		if startWordMarker == -1 && unicode.IsSpace(headRune) {
			continue
		}

		if startWordMarker == -1 {
			startWordMarker = head

			for _, quoteRune := range quoteRunes {
				if headRune == quoteRune {
					lastQuoteRune = quoteRune
				}
			}
		} else {
			if lastQuoteRune > 0 {
				if headRune == lastQuoteRune && lastRune != quoteEscapeRune {
					if removeQuotes {
						firstWordSansQuotes := startWordMarker + utf8.RuneLen(headRune)
						if head > firstWordSansQuotes {
							stringSansQuotes := strings.TrimSpace(string(commandLineBytes[firstWordSansQuotes:head]))
							tokenizedLines = append(tokenizedLines, stringSansQuotes)
						}
					} else {
						tokenizedLines = append(tokenizedLines, string(commandLineBytes[startWordMarker:head+width]))
					}

					startWordMarker = -1
					lastQuoteRune = 0
				}
			} else if unicode.IsSpace(headRune) {
				tokenizedLines = append(tokenizedLines, string(commandLineBytes[startWordMarker:head]))
				startWordMarker = -1
			}
		}

		lastRune = headRune
	}

	if lastQuoteRune > 0 {
		return nil, errors.Errorf("Unterminated quote sequence: %s", string(commandLineBytes[startWordMarker:]))
	}

	if startWordMarker != -1 {
		tokenizedLines = append(tokenizedLines, string(commandLineBytes[startWordMarker:]))
	}

	return tokenizedLines, nil
}

func CascadeErrorMessage(err1, err2 error) error {
	if err2 != nil {
		return errors.Errorf("%v\nThe error caused an additional error:\n%v", err1, err2)
	}

	return err1
}

func EscapeForShell(unsafeString string) string {
	return "'" + strings.ReplaceAll(unsafeString, `'`, `'"'"'`) + "'"
}

func PrettyPrintStringsForShell(args [][]string) string {
	output := ""

	for i, arg := range args {
		if len(arg) == 0 {
			continue
		}

		output = output + strings.Join(arg, " ")
		if i+1 < len(args) {
			output = output + " \\\n  "
		}
	}

	return output
}
