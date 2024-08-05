package parse

import (
	"strings"
)

type sourceMarker struct {
	lineContents    string
	sourceLineIndex int
}

type expandedSourceMarker struct {
	tokens          []token
	content         string
	comment         string
	sourceLineIndex int
	expanded        bool
}

func (e expandedSourceMarker) String() string {
	if len(e.tokens) == 0 {
		return ""
	}

	result := ""
	for _, content := range e.tokens {
		result += content.fatalContent
	}

	return result
}

func (e expandedSourceMarker) getTextContent() string {
	result := ""

	for _, token := range e.tokens {
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
		newExpandedTemplateLine := expandedSourceMarker{expanded: false, sourceLineIndex: expandedTemplateLine.sourceLineIndex}

		for _, token := range expandedTemplateLine.tokens {
			if token.tokenType != iterationType {
				newExpandedTemplateLine.tokens = append(newExpandedTemplateLine.tokens, token)
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
			newExpandedTemplateLine.tokens = append(newExpandedTemplateLine.tokens, newLineTokens...)

			if fatal != "" {
				aggregatedFatals = append(aggregatedFatals, aggregatedFatal{
					fatal:                     fatal,
					expandedTemplateLineIndex: newExpandedTemplateIdx,
				})
			}

			for newLineIndex, newLine := range newLines[1:] {
				newLineTokens, fatal := Tokenize(newLine, iterationType-1)
				newExpandedTemplateLines = append(newExpandedTemplateLines, newExpandedTemplateLine)

				newExpandedTemplateLine = expandedSourceMarker{expanded: true, sourceLineIndex: expandedTemplateLine.sourceLineIndex}
				newExpandedTemplateLine.tokens = append(newExpandedTemplateLine.tokens, newLineTokens...)

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

func (s *sectionedTemplate) expandTemplateLines(
	tokenize func(string) ([]token, bool, string),
	iterator func(t token) (string, string),
) {
	newExpandedTemplateLines := []expandedSourceMarker{}

	for _, expandedTemplateLine := range s.expandedTemplateLines {
		tokens, hasTokens, fatal := tokenize(expandedTemplateLine.content)

		if fatal != "" {
			s.setFatalMessage(fatal, expandedTemplateLine.sourceLineIndex)
			continue
		}

		if !hasTokens {
			newExpandedTemplateLines = append(newExpandedTemplateLines, expandedTemplateLine)
			continue
		}

		content := ""
		comment := ""

		for tokenIdx, token := range tokens {
			if token.tokenType == textToken {
				// Remove the escaping of `${ - because now it's ok to return
				// `${ and it'll be verbatim this from now on. So if a script
				// (or an env-var) contains that sequence it should not be erased
				// anymore.
				content += token.content
				continue
			}

			value, fatal := iterator(token)

			if fatal != "" {
				s.setFatalMessage(fatal, expandedTemplateLine.sourceLineIndex)
				continue
			}

			if s.hasFatalMessages() {
				// If there are errors the stuff below is busywork
				// Since we won't set any new expanded template lines
				// if there are fatals
				continue
			}

			value = strings.ReplaceAll(value, "\r\n", "\n")
			newLines := strings.Split(value, "\n")

			valueText, valueComment := splitTextOnComment(newLines[0])

			content += valueText
			comment = valueComment

			for _, newLine := range newLines[1:] {
				newExpandedTemplateLines = append(newExpandedTemplateLines, expandedSourceMarker{
					content:         content,
					comment:         valueComment,
					sourceLineIndex: expandedTemplateLine.sourceLineIndex,
					expanded:        true,
				})

				valueText, valueComment := splitTextOnComment(newLine)

				content = valueText
				comment = valueComment
			}

			if comment != "" {
				for _, restToken := range tokens[tokenIdx+1:] {
					comment += restToken.fatalContent
				}

				break
			}
		}

		newExpandedTemplateLines = append(newExpandedTemplateLines, expandedSourceMarker{
			content:         content,
			comment:         comment + expandedTemplateLine.comment,
			sourceLineIndex: expandedTemplateLine.sourceLineIndex,
			expanded:        true,
		})
	}

	if !s.hasFatalMessages() {
		s.expandedTemplateLines = newExpandedTemplateLines
	}
}

func newSectionedTemplate(rawTemplateString, filename string) *sectionedTemplate {
	rawTemplateLines := strings.Split(strings.ReplaceAll(rawTemplateString, "\r\n", "\n"), "\n")
	expandedTemplateLines := []expandedSourceMarker{}
	aggregatedFatals := []aggregatedFatal{}

	for sourceIndex, rawTemplateLine := range rawTemplateLines {
		tokens, fatal := Tokenize(rawTemplateLine, envVarToken)

		expandedTemplateLines = append(expandedTemplateLines, expandedSourceMarker{
			tokens:          tokens,
			sourceLineIndex: sourceIndex,
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

func newSectionedTemplate2(rawTemplateString, filename string) *sectionedTemplate {
	rawTemplateLines := strings.Split(strings.ReplaceAll(rawTemplateString, "\r\n", "\n"), "\n")

	expandedTemplateLines := []expandedSourceMarker{}

	for sourceIndex, rawTemplateLine := range rawTemplateLines {
		content, comment := splitTextOnComment(rawTemplateLine)

		expandedTemplateLines = append(expandedTemplateLines, expandedSourceMarker{
			content: content,
			comment: comment,

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
