package parse

import (
	"fmt"
	"regexp"
	"strings"
)

type captureResult struct {
	sectionHeaderLine sourceMarker
	sectionLines      []sourceMarker
}

const knownSectionHeaders = "host|query|headers|method|body|config|backend|backendoptions"

var knownSectionsRe = regexp.MustCompile(`^\[(` + knownSectionHeaders + `)\]$`)
var unescapeKnownSectionsRe = regexp.MustCompile(`^\\\[(` + knownSectionHeaders + `)\]$`)

func captureSection(sectionName string, template []sourceMarker, trim bool) (*captureResult, *fatalMarker) {
	var sectionLines []sourceMarker
	sectionHeaderLine := emptyLine
	capturing := false

	for _, templateLine := range template {
		lineContents := templateLine.lineContents
		trimmedLineContents := strings.TrimSpace(templateLine.lineContents)
		lowerCasedTrimmedLineContents := strings.ToLower(trimmedLineContents)

		if lowerCasedTrimmedLineContents == "["+strings.ToLower(sectionName)+"]" {
			if sectionHeaderLine != emptyLine {
				return nil, newFatalMarker(fmt.Sprintf("Several [%s] sections found on line %d and %d", sectionName, sectionHeaderLine.sourceLineIndex, templateLine.sourceLineIndex), emptyLine)
			}

			sectionHeaderLine = templateLine
			capturing = true
			continue
		}

		if knownSectionsRe.MatchString(lowerCasedTrimmedLineContents) {
			capturing = false
			continue
		}

		if unescapeKnownSectionsRe.MatchString(lowerCasedTrimmedLineContents) {
			lineContents = strings.Replace(lineContents, `\`, "", 1)
			trimmedLineContents = strings.Replace(trimmedLineContents, `\`, "", 1)
		}

		if capturing {
			if trim {
				templateLine.lineContents = trimmedLineContents
			} else {
				templateLine.lineContents = lineContents
			}

			sectionLines = append(sectionLines, templateLine)
		}
	}

	if sectionHeaderLine != emptyLine && len(sectionLines) == 0 {
		return nil, newFatalMarker("Empty ["+sectionName+"] line", sectionHeaderLine)
	}

	captureResult := &captureResult{
		sectionHeaderLine: sectionHeaderLine,
		sectionLines:      sectionLines,
	}

	return captureResult, nil
}
