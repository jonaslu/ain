package parse

import (
	"strings"
)

type sourceMarker struct {
	LineContents    string
	SourceLineIndex int
}

type expandedSourceMarker struct {
	Tokens          []token
	SourceLineIndex int
	expanded        bool
}

func (e expandedSourceMarker) String() string {
	if len(e.Tokens) == 0 {
		return ""
	}

	result := ""
	for _, content := range e.Tokens {
		result += content.fatalContent
	}

	return result
}

func (e expandedSourceMarker) getTextContent() string {
	result := ""

	for _, token := range e.Tokens {
		if token.tokenType == commentToken {
			break
		}

		result += token.content
	}

	return result
}

const (
	configSection         = "[config]"
	hostSection           = "[host]"
	querySection          = "[query]"
	headersSection        = "[headers]"
	methodSection         = "[method]"
	bodySection           = "[body]"
	backendSection        = "[backend]"
	backendOptionsSection = "[backendoptions]"
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
	// sourceMarker.SourceLineIndex points to the expandedTemplateLines slice
	sections map[string]*[]sourceMarker

	// sourceMarker.SourceLineIndex points to the rawTemplateLines slice
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

type aggregatedFatal struct {
	fatal                     string
	expandedTemplateLineIndex int
}

func (s *sectionedTemplate) iterate(iterationType tokenType, iterator func(token) (string, string)) {
	newExpandedTemplateLines := []expandedSourceMarker{}
	aggregatedFatals := []aggregatedFatal{}

	for expandedTemplateLineIdx, expandedTemplateLine := range s.expandedTemplateLines {
		newExpandedTemplateIdx := len(newExpandedTemplateLines)
		newExpandedTemplateLine := expandedSourceMarker{expanded: false, SourceLineIndex: expandedTemplateLine.SourceLineIndex}

		for _, token := range expandedTemplateLine.Tokens {
			if token.tokenType != iterationType {
				newExpandedTemplateLine.Tokens = append(newExpandedTemplateLine.Tokens, token)
				continue
			}

			newContentStr, fatal := iterator(token)
			if fatal != "" {
				// We'll use the old content, since the new wasn't valid
				s.setFatalMessage(fatal, expandedTemplateLineIdx)
				continue
			}

			newExpandedTemplateLine.expanded = true

			newContentStr = strings.ReplaceAll(newContentStr, "\r\n", "\n")
			newLines := strings.Split(newContentStr, "\n")

			newLineTokens, fatal := Tokenize(newLines[0], iterationType-1)
			newExpandedTemplateLine.Tokens = append(newExpandedTemplateLine.Tokens, newLineTokens...)

			if fatal != "" {
				aggregatedFatals = append(aggregatedFatals, aggregatedFatal{
					fatal:                     fatal,
					expandedTemplateLineIndex: newExpandedTemplateIdx,
				})
			}

			for newLineIndex, newLine := range newLines[1:] {
				newLineTokens, fatal := Tokenize(newLine, iterationType-1)
				newExpandedTemplateLines = append(newExpandedTemplateLines, newExpandedTemplateLine)

				newExpandedTemplateLine = expandedSourceMarker{expanded: true, SourceLineIndex: expandedTemplateLine.SourceLineIndex}
				newExpandedTemplateLine.Tokens = append(newExpandedTemplateLine.Tokens, newLineTokens...)

				if fatal != "" {
					aggregatedFatals = append(aggregatedFatals, aggregatedFatal{
						fatal: fatal,
						// + 1 because we've added at least one before splitting and adding more lines here
						expandedTemplateLineIndex: newExpandedTemplateIdx + newLineIndex + 1,
					})
				}
			}
		}

		newExpandedTemplateLines = append(newExpandedTemplateLines, newExpandedTemplateLine)
	}

	s.expandedTemplateLines = newExpandedTemplateLines

	for _, aggregatedFatal := range aggregatedFatals {
		s.setFatalMessage(aggregatedFatal.fatal, aggregatedFatal.expandedTemplateLineIndex)
	}
}

func newSectionedTemplate(rawTemplateString, filename string) *sectionedTemplate {
	rawTemplateLines := strings.Split(strings.ReplaceAll(rawTemplateString, "\r\n", "\n"), "\n")
	expandedTemplateLines := []expandedSourceMarker{}
	aggregatedFatals := []aggregatedFatal{}

	for sourceIndex, rawTemplateLine := range rawTemplateLines {
		tokens, fatal := Tokenize(rawTemplateLine, envVarToken)

		expandedTemplateLines = append(expandedTemplateLines, expandedSourceMarker{
			Tokens:          tokens,
			SourceLineIndex: sourceIndex,
			expanded:        false,
		})

		if fatal != "" {
			aggregatedFatals = append(aggregatedFatals, aggregatedFatal{
				fatal:                     fatal,
				expandedTemplateLineIndex: sourceIndex,
			})
		}
	}

	sectionedTemplate := sectionedTemplate{
		sections:              map[string]*[]sourceMarker{},
		expandedTemplateLines: expandedTemplateLines,
		rawTemplateLines:      rawTemplateLines,
		filename:              filename,
	}

	for _, aggregatedFatal := range aggregatedFatals {
		sectionedTemplate.setFatalMessage(aggregatedFatal.fatal, aggregatedFatal.expandedTemplateLineIndex)
	}

	return &sectionedTemplate
}
