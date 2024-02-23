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

var isCommentOrWhitespaceRegExp = regexp.MustCompile(`^\s*#|^\s*$`)

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

	// !! TODO !! We now know the sourceLineIndex where multiple headings
	// are defined so we can print this more nicely
	for heading, headingSourceLineIndex := range headingDefinitionSourceLines {
		if len(headingSourceLineIndex) > 1 {
			s.fatals = append(
				s.fatals,
				fmt.Sprintf("Several %s sections found on line %d and %d", heading, headingSourceLineIndex[0]+1, headingSourceLineIndex[1]+1))
		}
	}
}

func containsHeader(sectionHeading string, wantedSectionHeadings []string) bool {
	for _, val := range wantedSectionHeadings {
		if sectionHeading == val {
			return true
		}
	}

	return false
}

func (s *sectionedTemplate) setCapturedSections(wantedSectionHeadings ...string) {
	capturedSections := []capturedSection{}
	var currentSectionHeader string
	var currentSectionLines *[]sourceMarker

	for expandedSourceIndex, expandedTemplateLine := range s.expandedTemplateLines {
		lineContents := expandedTemplateLine.LineContents

		if currentSectionHeader != bodySection && isCommentOrWhitespaceRegExp.MatchString(lineContents) {
			continue
		}

		trailingCommentsRemoved, _, _ := strings.Cut(lineContents, "#")

		if sectionHeading := getSectionHeading(trailingCommentsRemoved); sectionHeading != "" {

			if currentSectionHeader == bodySection {
				firstNonEmptyLine := 0
				for ; firstNonEmptyLine < len(*currentSectionLines); firstNonEmptyLine++ {
					if !isCommentOrWhitespaceRegExp.MatchString((*currentSectionLines)[firstNonEmptyLine].LineContents) {
						break
					}
				}

				lastNonEnptyLine := len(*currentSectionLines) - 1
				for ; lastNonEnptyLine > firstNonEmptyLine; lastNonEnptyLine-- {
					if !isCommentOrWhitespaceRegExp.MatchString((*currentSectionLines)[lastNonEnptyLine].LineContents) {
						break
					}
				}

				*currentSectionLines = (*currentSectionLines)[firstNonEmptyLine : lastNonEnptyLine+1]
			}

			if !containsHeader(sectionHeading, wantedSectionHeadings) {
				currentSectionLines = nil
				currentSectionHeader = ""
				continue
			}

			currentSectionLines = &[]sourceMarker{}
			currentSectionHeader = sectionHeading
			capturedSections = append(capturedSections, capturedSection{
				heading:                sectionHeading,
				headingSourceLineIndex: expandedSourceIndex,
				sectionLines:           currentSectionLines,
			})

			continue
		}

		if currentSectionLines == nil {
			continue
		}

		if unescapeKnownSectionsRe.MatchString(trailingCommentsRemoved) {
			trailingCommentsRemoved = strings.Replace(trailingCommentsRemoved, `\`, "", 1)
		}

		var lineContentsTrimmed string
		if currentSectionHeader == bodySection {
			lineContentsTrimmed = strings.TrimRightFunc(trailingCommentsRemoved, func(r rune) bool { return unicode.IsSpace(r) })
		} else {
			lineContentsTrimmed = strings.TrimSpace(trailingCommentsRemoved)
		}

		sourceMarker := sourceMarker{
			LineContents:    lineContentsTrimmed,
			SourceLineIndex: expandedSourceIndex,
		}

		*currentSectionLines = append(*currentSectionLines, sourceMarker)
	}

	if s.checkValidHeadings(capturedSections); s.hasFatalMessages() {
		return
	}

	for _, capturedSection := range capturedSections {
		s.sections[capturedSection.heading] = capturedSection.sectionLines
	}
}
