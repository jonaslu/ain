package parse

import (
	"fmt"
	"strings"
	"unicode"
)

type capturedSection struct {
	heading                string
	headingSourceLineIndex int
	sectionLines           *[]sourceMarker
}

func getSectionHeading(templateLineTextTrimmed string) string {
	templateLineTextTrimmedLower := strings.ToLower(templateLineTextTrimmed)
	for _, knownSectionHeader := range allSectionHeaders {
		if templateLineTextTrimmedLower == knownSectionHeader {
			return knownSectionHeader
		}
	}

	return ""
}

func (s *sectionedTemplate) checkValidHeadings(capturedSections []capturedSection) {
	// Keeps "header": [1,5,7] <- Name of heading and on what lines in the file
	headingDefinitionSourceLines := map[string][]int{}

	for _, capturedSection := range capturedSections {
		if len(*capturedSection.sectionLines) == 0 {
			// !! TODO !! Can I use capturedSectionLine or so
			s.setFatalMessage(fmt.Sprintf("Empty %s section", capturedSection.heading), capturedSection.headingSourceLineIndex)
		}

		headingDefinitionSourceLines[capturedSection.heading] = append(headingDefinitionSourceLines[capturedSection.heading], capturedSection.headingSourceLineIndex)
	}

	for heading, headingSourceLineIndexes := range headingDefinitionSourceLines {
		if len(headingSourceLineIndexes) == 1 {
			continue
		}

		for _, headingSourceLineIndex := range headingSourceLineIndexes[1:] {
			s.setFatalMessage(fmt.Sprintf("Section %s on line %d redeclared", heading, headingSourceLineIndexes[0]+1), headingSourceLineIndex)
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
		if (*currentSectionLines)[firstNonEmptyLine].lineContents != "" {
			break
		}
	}

	lastNonEmptyLine := len(*currentSectionLines) - 1
	for ; lastNonEmptyLine > firstNonEmptyLine; lastNonEmptyLine-- {
		if (*currentSectionLines)[lastNonEmptyLine].lineContents != "" {
			break
		}
	}

	*currentSectionLines = (*currentSectionLines)[firstNonEmptyLine : lastNonEmptyLine+1]
}

func unescapeSectionHeading(templateLineTextTrimmed, templateLineText string) string {
	templateLineTextTrimmedLower := strings.ToLower(templateLineTextTrimmed)

	// !! DEPRECATE !! Old way (e g  \[Body])
	if strings.HasPrefix(templateLineTextTrimmed, `\`) {
		for _, knownSectionHeader := range allSectionHeaders {
			if templateLineTextTrimmedLower == `\`+knownSectionHeader {
				return strings.Replace(templateLineText, `\`, "", 1)
			}
		}
	}

	if strings.HasPrefix(templateLineTextTrimmed, "`") {
		for _, knownSectionHeader := range allSectionHeaders {
			if templateLineTextTrimmedLower == "`"+knownSectionHeader {
				return strings.Replace(templateLineText, "`", "", 1)
			}
		}
	}

	if strings.HasPrefix(templateLineTextTrimmed, "\\`") {
		for _, knownSectionHeader := range allSectionHeaders {
			if templateLineTextTrimmedLower == "\\`"+knownSectionHeader {
				return strings.Replace(templateLineText, "\\`", "`", 1)
			}
		}
	}

	return templateLineText
}

func (s *sectionedTemplate) setCapturedSections(wantedSectionHeadings ...string) {
	capturedSections := []capturedSection{}

	var currentSectionHeader string
	var currentSectionLines *[]sourceMarker

	for expandedSourceIndex, _ := range s.expandedTemplateLines {
		expandedTemplateLine := &s.expandedTemplateLines[expandedSourceIndex]

		templateLineText := expandedTemplateLine.getTextContent()
		templateLineTextTrimmed := strings.TrimSpace(templateLineText)

		if currentSectionHeader != "" {
			expandedTemplateLine.consumed = true
		}

		// Discard empty lines, except if it's the [Body] section
		if currentSectionHeader != bodySection && templateLineTextTrimmed == "" {
			continue
		}

		if sectionHeading := getSectionHeading(templateLineTextTrimmed); sectionHeading != "" {
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

			expandedTemplateLine.consumed = true

			continue
		}

		// Not a section we're interested in
		if currentSectionLines == nil {
			continue
		}

		templateLineText = unescapeSectionHeading(templateLineTextTrimmed, templateLineText)

		sourceMarker := sourceMarker{
			sourceLineIndex: expandedSourceIndex,
		}

		if currentSectionHeader == bodySection {
			sourceMarker.lineContents = strings.TrimRightFunc(templateLineText, func(r rune) bool { return unicode.IsSpace(r) })
		} else {
			sourceMarker.lineContents = strings.TrimSpace(templateLineText)
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
