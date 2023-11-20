package utils

import (
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

const quoteEscapeRune = '\\'

var quoteRunes = [...]rune{'"', '\''}

const unterminatedQuoteErrorMessageContext = 3

func TokenizeLine(commandLine string) ([]string, error) {
	var tokenizedLines []string
	var lastQuoteRune rune
	var lastQuotePos int

	commandLineRune := []rune(commandLine)

	var builder strings.Builder
	builder.Grow(len(commandLine))

NextRune:
	for i := 0; i < len(commandLineRune); i++ {
		headRune := commandLineRune[i]

		// Nothing has been collected, discard spaces
		if lastQuoteRune == 0 && unicode.IsSpace(headRune) && builder.Len() == 0 {
			continue
		}

		// Quoting is turned on
		if lastQuoteRune > 0 {
			// Escaped quote \" - write only the quote and carry on
			if headRune == quoteEscapeRune && i < len(commandLineRune)-1 && commandLineRune[i+1] == lastQuoteRune {
				builder.WriteRune(lastQuoteRune)
				i = i + 1
				continue
			}

			// Turns quoting off
			if headRune == lastQuoteRune {
				lastQuoteRune = 0
				continue
			}

			builder.WriteRune(headRune)
			continue
		}

		// Quoting not turned on, look for any escaped quote
		if headRune == quoteEscapeRune && i < len(commandLineRune)-1 {
			for _, quoteRune := range quoteRunes {
				if commandLineRune[i+1] == quoteRune {
					builder.WriteRune(quoteRune)
					i = i + 1
					continue NextRune
				}
			}
		}

		// Check for start of quoting
		for _, quoteRune := range quoteRunes {
			if headRune == quoteRune {
				lastQuoteRune = quoteRune
				lastQuotePos = i
				continue NextRune
			}
		}

		// We're not quoting and we are on a word boundary
		if unicode.IsSpace(headRune) && lastQuoteRune == 0 {
			tokenizedLines = append(tokenizedLines, builder.String())
			builder.Reset()
			continue
		}

		builder.WriteRune(headRune)
	}

	if lastQuoteRune > 0 {
		var context string
		preContext := lastQuotePos - unterminatedQuoteErrorMessageContext

		if preContext < 1 {
			preContext = 0
		} else {
			context = "..."
		}

		subContext := lastQuotePos + unterminatedQuoteErrorMessageContext + 1
		if subContext >= len(commandLine) {
			subContext = len(commandLine)
		}

		context = context + commandLine[preContext:subContext]

		if lastQuotePos+unterminatedQuoteErrorMessageContext < len(commandLine)-1 {
			context = context + "..."
		}

		return nil, errors.Errorf("Unterminated quote sequence: %s", context)
	}

	if builder.Len() > 0 {
		tokenizedLines = append(tokenizedLines, builder.String())
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
