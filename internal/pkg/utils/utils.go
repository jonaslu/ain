package utils

import (
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
)

const quoteEscapeRune = '\\'

var quoteRunes = [...]rune{'"', '\''}

// Thi func is copied from the go source-code. I wish it was an exported
// method so I didn't have to. Here's the
// link to it's licence: https://golang.org/LICENSE
func isSpace(r rune) bool {
	if r <= '\u00FF' {
		switch r {
		case ' ', '\t', '\n', '\v', '\f', '\r':
			return true
		case '\u0085', '\u00A0':
			return true
		}
		return false
	}
	if '\u2000' <= r && r <= '\u200a' {
		return true
	}
	switch r {
	case '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000':
		return true
	}
	return false
}

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

		if startWordMarker == -1 && isSpace(headRune) {
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
			} else if isSpace(headRune) {
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
