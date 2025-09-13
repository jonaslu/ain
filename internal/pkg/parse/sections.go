package parse

import (
	"fmt"
	"os"
	"strings"
)

type sourceMarker struct {
	lineContents    string
	sourceLineIndex int
}

type expandedSourceMarker struct {
	content         string
	fatalContent    string
	comment         string
	sourceLineIndex int
	expanded        bool
	consumed        bool
}

func (e expandedSourceMarker) String() string {
	return e.fatalContent + e.comment
}

func (e expandedSourceMarker) getTextContent() string {
	textContent := strings.ReplaceAll(e.content, "`"+commentPrefix, commentPrefix)

	if e.comment != "" && strings.HasSuffix(textContent, "\\`") {
		textContent = strings.TrimSuffix(textContent, "\\`") + "`"
	}

	return textContent
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

func splitValueOnNewlines(value string, currentLine expandedSourceMarker) (newExpandedTemplateLines []expandedSourceMarker) {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	newLines := strings.Split(value, "\n")

	valueText, valueComment := splitTextOnComment(newLines[0])

	currentLine.content += valueText
	currentLine.fatalContent += valueText
	currentLine.comment = valueComment

	for _, newLine := range newLines[1:] {
		newExpandedTemplateLines = append(newExpandedTemplateLines, currentLine)
		currentLine = expandedSourceMarker{sourceLineIndex: currentLine.sourceLineIndex, expanded: true}

		valueText, valueComment := splitTextOnComment(newLine)

		currentLine.content = valueText
		currentLine.fatalContent = valueText
		currentLine.comment = valueComment
	}

	return append(newExpandedTemplateLines, currentLine)
}

func (s *sectionedTemplate) processLineTokens(
	tokens,
	fatalTokens []token,
	tokenIterator func(t token) (string, string),
	previousLine expandedSourceMarker,
) []expandedSourceMarker {
	currentLine := expandedSourceMarker{sourceLineIndex: previousLine.sourceLineIndex, expanded: previousLine.expanded}
	newExpandedTemplateLines := []expandedSourceMarker{}

	// Range over the lines tokens
	for tokenIdx, token := range tokens {
		if token.tokenType == textToken {
			// Remove the escaping of `${ - because now it's ok to return
			// `${ and it'll be verbatim this from now on. So if a script
			// (or an env-var) contains that sequence it should not be erased
			// anymore.
			currentLine.content += token.content
			currentLine.fatalContent += fatalTokens[tokenIdx].fatalContent

			continue
		}

		value, fatal := tokenIterator(token)
		if fatal != "" {
			s.setFatalMessage(fatal, previousLine.sourceLineIndex)
			continue
		}

		if s.hasFatalMessages() {
			// Fatals relates to the current expanded lines,
			// and not the new we're making. Avoid the computation
			// below but keep iterating over tokens so
			// we report all fatals on this line
			continue
		}

		currentLine.expanded = true

		splitOnLineMarkers := splitValueOnNewlines(value, currentLine)
		newExpandedTemplateLines = append(newExpandedTemplateLines, splitOnLineMarkers[:len(splitOnLineMarkers)-1]...)
		currentLine = splitOnLineMarkers[len(splitOnLineMarkers)-1]

		if currentLine.comment != "" {
			for _, restToken := range tokens[tokenIdx+1:] {
				currentLine.comment += restToken.fatalContent
			}
			break
		}
	}

	currentLine.comment += previousLine.comment
	newExpandedTemplateLines = append(newExpandedTemplateLines, currentLine)

	return newExpandedTemplateLines
}

func (s *sectionedTemplate) expandTemplateLines(
	tokenizer func(string) ([]token, string),
	tokenIterator func(t token) (string, string),
) {
	newExpandedTemplateLines := []expandedSourceMarker{}

	for _, expandedTemplateLine := range s.expandedTemplateLines {
		if expandedTemplateLine.consumed {
			newExpandedTemplateLines = append(newExpandedTemplateLines, expandedTemplateLine)
			continue
		}

		tokens, fatal := tokenizer(expandedTemplateLine.content)
		if fatal != "" {
			s.setFatalMessage(fatal, expandedTemplateLine.sourceLineIndex)
			continue
		}

		fatalTokens, fatal := tokenizer(expandedTemplateLine.fatalContent)
		if fatal != "" {
			fmt.Fprintf(os.Stderr, "Internal error tokenizing fatals: %s\n", fatal)
			os.Exit(1)
		}

		// One token might return several lines
		expandedLinesFromTokens := s.processLineTokens(
			tokens,
			fatalTokens,
			tokenIterator,
			expandedTemplateLine,
		)

		newExpandedTemplateLines = append(newExpandedTemplateLines, expandedLinesFromTokens...)
	}

	if !s.hasFatalMessages() {
		s.expandedTemplateLines = newExpandedTemplateLines
	}
}

func newSectionedTemplate(rawTemplateString, filename string) *sectionedTemplate {
	rawTemplateLines := strings.Split(strings.ReplaceAll(rawTemplateString, "\r\n", "\n"), "\n")

	expandedTemplateLines := []expandedSourceMarker{}

	for sourceIndex, rawTemplateLine := range rawTemplateLines {
		content, comment := splitTextOnComment(rawTemplateLine)

		expandedTemplateLines = append(expandedTemplateLines, expandedSourceMarker{
			content:      content,
			fatalContent: content,
			comment:      comment,

			sourceLineIndex: sourceIndex,
			expanded:        false,
		})
	}

	sectionedTemplate := sectionedTemplate{
		sections:              map[string]*[]sourceMarker{},
		expandedTemplateLines: expandedTemplateLines,
		rawTemplateLines:      rawTemplateLines,
		filename:              filename,
	}

	return &sectionedTemplate
}
