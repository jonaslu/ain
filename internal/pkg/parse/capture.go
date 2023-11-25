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

var emptyOutputLineRe = regexp.MustCompile(`^\s*$`)

var removeTrailingCommendRegExp = regexp.MustCompile("#.*$")
var isCommentOrWhitespaceRegExp = regexp.MustCompile(`^\s*#|^\s*$`)

func (s *sectionedTemplate) splitAndTrimSection(sectionHeader string) {
	sourceMarkers := s.getNamedSection(sectionHeader)
	replacedSection := []sourceMarker{}

	for _, templateLine := range *sourceMarkers {
		multilineOutput := strings.Split(strings.ReplaceAll(templateLine.LineContents, "\r\n", "\n"), "\n")

		for _, lineOutput := range multilineOutput {
			if sectionHeader != bodySection {
				if emptyOutputLineRe.MatchString(lineOutput) {
					continue
				}

				templateLine.LineContents = strings.TrimSpace(lineOutput)
			} else {
				templateLine.LineContents = lineOutput
			}

			replacedSection = append(replacedSection, templateLine)
		}
	}

	s.sections[sectionHeader] = &replacedSection
}

func (s *sectionedTemplate) setCapturedSections(capturedSections []capturedSection) {
	for _, capturedSection := range capturedSections {
		s.sections[capturedSection.heading] = capturedSection.sectionLines
	}
}

func getSectionHeading(rawTemplateLine string) string {
	matchedLine := knownSectionsRe.FindStringSubmatch(rawTemplateLine)

	if len(matchedLine) == 2 {
		return strings.ToLower(matchedLine[1])
	}

	return ""
}

func checkValidHeadings(capturedSections []capturedSection, sections *sectionedTemplate) {
	// Keeps "header": [1,5,7] <- Name of heading and on what lines in the file
	headingDefinitionSourceLines := map[string][]int{}

	for _, capturedSection := range capturedSections {
		if len(*capturedSection.sectionLines) == 0 {
			sections.setFatalMessage(fmt.Sprintf("Empty %s section", capturedSection.heading), capturedSection.headingSourceLineIndex)
		}

		headingDefinitionSourceLines[capturedSection.heading] = append(headingDefinitionSourceLines[capturedSection.heading], capturedSection.headingSourceLineIndex)
	}

	// !! TODO !! We now know the sourceLineIndex where multiple headings
	// are defined so we can print this more nicely
	for heading, headingSourceLineIndex := range headingDefinitionSourceLines {
		if len(headingSourceLineIndex) > 1 {
			sections.fatals = append(
				sections.fatals,
				fmt.Sprintf("Several %s sections found on line %d and %d", heading, headingSourceLineIndex[0]+1, headingSourceLineIndex[1]+1))
		}
	}
}

func getCapturedSections(rawTemplateLines []string) ([]capturedSection, bool) {
	templateEmpty := true
	capturedSections := []capturedSection{}
	currentSectionLines := &[]sourceMarker{}

	for sourceIndex, rawTemplateLine := range rawTemplateLines {
		if isCommentOrWhitespaceRegExp.MatchString(rawTemplateLine) {
			continue
		}

		trailingCommentsRemoved := removeTrailingCommendRegExp.ReplaceAllString(rawTemplateLine, "")

		if sectionHeading := getSectionHeading(trailingCommentsRemoved); sectionHeading != "" {
			currentSectionLines = &[]sourceMarker{}
			capturedSections = append(capturedSections, capturedSection{
				heading:                sectionHeading,
				headingSourceLineIndex: sourceIndex,
				sectionLines:           currentSectionLines,
			})

			templateEmpty = false

			continue
		}

		if unescapeKnownSectionsRe.MatchString(trailingCommentsRemoved) {
			trailingCommentsRemoved = strings.Replace(trailingCommentsRemoved, `\`, "", 1)
		}

		sourceMarker := sourceMarker{
			LineContents:    strings.TrimRightFunc(trailingCommentsRemoved, func(r rune) bool { return unicode.IsSpace(r) }),
			SourceLineIndex: sourceIndex,
		}

		*currentSectionLines = append(*currentSectionLines, sourceMarker)
	}

	return capturedSections, templateEmpty
}
