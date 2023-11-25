package parse

import (
	"strings"
)

type sourceMarker struct {
	LineContents    string
	SourceLineIndex int
}

const (
	configSection         = "config"
	hostSection           = "host"
	querySection          = "query"
	headersSection        = "headers"
	methodSection         = "method"
	bodySection           = "body"
	backendSection        = "backend"
	backendOptionsSection = "backendoptions"
	// As above, so below
	// If you add one here then add it to the slice below.
	// AND IF
	// it should be included when capturing executables (i e not Config
	// as it's parsed before running executables) add it to the
	// second slice below
)

var allSectionHeaders = []string{
	configSection,
	hostSection,
	querySection,
	headersSection,
	methodSection,
	bodySection,
	backendSection,
	backendOptionsSection,
}

var sectionsAllowingExecutables = []string{
	hostSection,
	querySection,
	headersSection,
	methodSection,
	bodySection,
	backendSection,
	backendOptionsSection,
}

type sectionedTemplate struct {
	sections map[string]*[]sourceMarker

	fatals []string

	rawTemplateLines []string
	filename         string
}

func (s *sectionedTemplate) getNamedSection(sectionHeader string) *[]sourceMarker {
	if section, exists := s.sections[sectionHeader]; exists {
		return section
	}

	return &[]sourceMarker{}
}

func newSectionedTemplate(rawTemplateString, filename string) *sectionedTemplate {
	rawTemplateLines := strings.Split(strings.ReplaceAll(rawTemplateString, "\r\n", "\n"), "\n")
	sectionedTemplate := sectionedTemplate{
		sections:         map[string]*[]sourceMarker{},
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

	if !sectionedTemplate.hasFatalMessages() {
		sectionedTemplate.setCapturedSections(capturedSections)
	}

	return &sectionedTemplate
}
