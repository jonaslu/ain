package main

import (
	"strings"
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

func (s *SectionedTemplate) GetNamedSection(sectionHeader string) *[]SourceMarker {
	if section, exists := s.sections[sectionHeader]; exists {
		return section
	}

	return &[]SourceMarker{}
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
