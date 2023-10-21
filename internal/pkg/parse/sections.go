package parse

import (
	"strings"
)

type sourceMarker struct {
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
