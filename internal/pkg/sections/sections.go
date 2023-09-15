package main

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type SourceMarker struct {
	LineContents    string
	SourceLineIndex int
}

const (
	ConfigSection         = "config"
	HostSection           = "host"
	QuerySection          = "query"
	HeadersSection        = "headers"
	MethodSection         = "method"
	BodySection           = "body"
	BackendSection        = "backend"
	BackendOptionsSection = "backendoptions"
	DefaultVarsSection    = "defaultvars"
)

type SectionedTemplate struct {
	sections map[string]*[]SourceMarker

	fatals []string

	rawTemplateLines []string
	filename         string
}

type capturedSection struct {
	heading                string
	headingSourceLineIndex int
	sectionLines           *[]SourceMarker
}

var allSectionHeaders = []string{
	ConfigSection,
	HostSection,
	QuerySection,
	HeadersSection,
	MethodSection,
	BodySection,
	BackendSection,
	BackendOptionsSection,
	DefaultVarsSection,
}

var knownSectionHeadersStr = strings.Join(allSectionHeaders, "|")

var knownSectionsRe = regexp.MustCompile(`(?i)^\s*\[(` + knownSectionHeadersStr + `)\]\s*$`)
var unescapeKnownSectionsRe = regexp.MustCompile(`(?i)^\s*\\\[(` + knownSectionHeadersStr + `)\]\s*$`)

var removeTrailingCommendRegExp = regexp.MustCompile("#.*$")
var isCommentOrWhitespaceRegExp = regexp.MustCompile(`^\s*#|^\s*$`)

func trimSourceMarkerLines(sourceMarkers *[]SourceMarker) *[]SourceMarker {
	for idx := range *sourceMarkers {
		sourceMarker := &(*sourceMarkers)[idx]
		sourceMarker.LineContents = strings.TrimSpace(sourceMarker.LineContents)
	}

	return sourceMarkers
}

func (s *SectionedTemplate) setCapturedSections(capturedSections []capturedSection) {
	for _, capturedSection := range capturedSections {
		if capturedSection.heading == BodySection {
			s.sections[capturedSection.heading] = capturedSection.sectionLines
		} else {
			s.sections[capturedSection.heading] = trimSourceMarkerLines(capturedSection.sectionLines)
		}
	}
}

func getSectionHeading(rawTemplateLine string) string {
	matchedLine := knownSectionsRe.FindStringSubmatch(rawTemplateLine)

	if len(matchedLine) == 2 {
		return strings.ToLower(matchedLine[1])
	}

	return ""
}

func checkValidHeadings(capturedSections []capturedSection, sections *SectionedTemplate) {
	// Keeps "header": [1,5,7] <- Name of heading and on what lines in the file
	headingDefinitionSourceLines := map[string][]int{}

	for _, capturedSection := range capturedSections {
		if len(*capturedSection.sectionLines) == 0 {
			sections.SetFatalMessage(fmt.Sprintf("Empty %s section", capturedSection.heading), capturedSection.headingSourceLineIndex)
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
	currentSectionLines := &[]SourceMarker{}

	for sourceIndex, rawTemplateLine := range rawTemplateLines {
		if isCommentOrWhitespaceRegExp.MatchString(rawTemplateLine) {
			continue
		}

		trailingCommentsRemoved := removeTrailingCommendRegExp.ReplaceAllString(rawTemplateLine, "")

		if sectionHeading := getSectionHeading(trailingCommentsRemoved); sectionHeading != "" {
			currentSectionLines = &[]SourceMarker{}
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

		sourceMarker := SourceMarker{
			LineContents:    strings.TrimRightFunc(trailingCommentsRemoved, func(r rune) bool { return unicode.IsSpace(r) }),
			SourceLineIndex: sourceIndex,
		}

		*currentSectionLines = append(*currentSectionLines, sourceMarker)
	}

	return capturedSections, templateEmpty
}

func NewSections(rawTemplateString, filename string) *SectionedTemplate {
	rawTemplateLines := strings.Split(strings.ReplaceAll(rawTemplateString, "\r\n", "\n"), "\n")
	sectionedTemplate := SectionedTemplate{
		sections:         map[string]*[]SourceMarker{},
		rawTemplateLines: rawTemplateLines,
		filename:         filename,
	}

	capturedSections, templateEmpty := getCapturedSections(rawTemplateLines)

	if templateEmpty {
		// !! TODO !! Change to no valid headings found, it can be full of stuff
		sectionedTemplate.fatals = []string{"Cannot process empty template"}
	} else {
		checkValidHeadings(capturedSections, &sectionedTemplate)
	}

	if !sectionedTemplate.HasFatalMessages() {
		sectionedTemplate.setCapturedSections(capturedSections)
	}

	return &sectionedTemplate
}
