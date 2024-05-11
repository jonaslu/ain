package parse

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type capturedSection struct {
	heading                string
	headingSourceLineIndex int
	sectionLines           *[]sourceMarker
}

var knownSectionHeadersStr = strings.Join(allSectionHeaders, "|")

var knownSectionsRe = regexp.MustCompile(`(?i)^\s*\[(` + knownSectionHeadersStr + `)\]\s*$`)
var unescapeKnownSectionsRe = regexp.MustCompile(`(?i)^\s*\\\[(` + knownSectionHeadersStr + `)\]\s*$`)

func getSectionHeading(rawTemplateLine string) string {
	matchedLine := knownSectionsRe.FindStringSubmatch(rawTemplateLine)

	if len(matchedLine) == 2 {
		return strings.ToLower(matchedLine[1])
	}

	return ""
}

func (s *sectionedTemplate) checkValidHeadings(capturedSections []capturedSection) {
	// Keeps "header": [1,5,7] <- Name of heading and on what lines in the file
	headingDefinitionSourceLines := map[string][]int{}

	for _, capturedSection := range capturedSections {
		if len(*capturedSection.sectionLines) == 0 {
			s.setFatalMessage(fmt.Sprintf("Empty %s section", capturedSection.heading), capturedSection.headingSourceLineIndex)
		}

		headingDefinitionSourceLines[capturedSection.heading] = append(headingDefinitionSourceLines[capturedSection.heading], capturedSection.headingSourceLineIndex)
	}

	for heading, headingSourceLineIndexes := range headingDefinitionSourceLines {
		if len(headingSourceLineIndexes) == 1 {
			continue
		}

		for _, headingSourceLineIndex := range headingSourceLineIndexes[1:] {
			s.setFatalMessage(fmt.Sprintf("Section %s on line %d redeclared", heading, headingSourceLineIndexes[0]), headingSourceLineIndex)
		}
	}
}

func containsSectionHeader(sectionHeading string, wantedSectionHeadings []string) bool {
	for _, val := range wantedSectionHeadings {
		if sectionHeading == val {
			return true
		}
	}

	return false
}

func compactBodySection(currentSectionLines *[]sourceMarker) {
	firstNonEmptyLine := 0
	for ; firstNonEmptyLine < len(*currentSectionLines); firstNonEmptyLine++ {
		if (*currentSectionLines)[firstNonEmptyLine].LineContents != "" {
			break
		}
	}

	lastNonEmptyLine := len(*currentSectionLines) - 1
	for ; lastNonEmptyLine > firstNonEmptyLine; lastNonEmptyLine-- {
		if (*currentSectionLines)[lastNonEmptyLine].LineContents != "" {
			break
		}
	}

	*currentSectionLines = (*currentSectionLines)[firstNonEmptyLine : lastNonEmptyLine+1]
}

func (s *sectionedTemplate) setCapturedSections(wantedSectionHeadings ...string) {
	capturedSections := []capturedSection{}

	var currentSectionHeader string
	var currentSectionLines *[]sourceMarker

	for expandedSourceIndex, expandedTemplateLine := range s.expandedTemplateLines {
		templateLineText := expandedTemplateLine.getTextContent()

		// Discard empty lines, except if it's the [Body] section
		if currentSectionHeader != bodySection && strings.TrimSpace(templateLineText) == "" {
			continue
		}

		if sectionHeading := getSectionHeading(templateLineText); sectionHeading != "" {
			// Compact [Body] section
			if currentSectionHeader == bodySection {
				compactBodySection(currentSectionLines)
			}

			if !containsSectionHeader(sectionHeading, wantedSectionHeadings) {
				currentSectionLines = nil
				currentSectionHeader = ""
				continue
			}

			currentSectionHeader = sectionHeading
			currentSectionLines = &[]sourceMarker{}

			capturedSections = append(capturedSections, capturedSection{
				heading:                sectionHeading,
				headingSourceLineIndex: expandedSourceIndex,
				sectionLines:           currentSectionLines,
			})

			continue
		}

		// Not a section we're interested in
		if currentSectionLines == nil {
			continue
		}

		if unescapeKnownSectionsRe.MatchString(templateLineText) {
			templateLineText = strings.Replace(templateLineText, `\`, "", 1)
		}

		var templateLineTextTrimmed string
		if currentSectionHeader == bodySection {
			templateLineTextTrimmed = strings.TrimRightFunc(templateLineText, func(r rune) bool { return unicode.IsSpace(r) })
		} else {
			templateLineTextTrimmed = strings.TrimSpace(templateLineText)
		}

		sourceMarker := sourceMarker{
			LineContents:    templateLineTextTrimmed,
			SourceLineIndex: expandedSourceIndex,
		}

		*currentSectionLines = append(*currentSectionLines, sourceMarker)
	}

	if currentSectionHeader == bodySection {
		compactBodySection(currentSectionLines)
	}

	if s.checkValidHeadings(capturedSections); s.hasFatalMessages() {
		return
	}

	for _, capturedSection := range capturedSections {
		s.sections[capturedSection.heading] = capturedSection.sectionLines
	}
}
