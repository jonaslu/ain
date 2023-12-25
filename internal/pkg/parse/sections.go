package parse

import (
	"strings"
)

type sourceMarker struct {
	LineContents    string
	SourceLineIndex int
}

type expandedSourceMarker struct {
	sourceMarker
	expanded bool
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
	// sourceMarker.LineContents points to the expandedTemplateLines slice
	sections map[string]*[]sourceMarker

	// sourceMarker.LineContents points to the rawTemplateLines slice
	expandedTemplateLines []expandedSourceMarker
	rawTemplateLines      []string

	filename string
	fatals   []string
}

func (s *sectionedTemplate) getNamedSection(sectionHeader string) *[]sourceMarker {
	if section, exists := s.sections[sectionHeader]; exists {
		return section
	}

	return &[]sourceMarker{}
}

func newSectionedTemplate(rawTemplateString, filename string) *sectionedTemplate {
	rawTemplateLines := strings.Split(strings.ReplaceAll(rawTemplateString, "\r\n", "\n"), "\n")
	expandedTemplateLines := []expandedSourceMarker{}

	for sourceIndex, rawTemplateLine := range rawTemplateLines {
		expandedTemplateLines = append(expandedTemplateLines, expandedSourceMarker{
			sourceMarker: sourceMarker{
				LineContents:    rawTemplateLine,
				SourceLineIndex: sourceIndex,
			},
		})
	}

	sectionedTemplate := sectionedTemplate{
		sections:              map[string]*[]sourceMarker{},
		expandedTemplateLines: expandedTemplateLines,
		rawTemplateLines:      rawTemplateLines,
		filename:              filename,
	}

	if len(expandedTemplateLines) == 0 {
		sectionedTemplate.fatals = []string{"Cannot process empty template"}
		return &sectionedTemplate
	}

	return &sectionedTemplate
}
