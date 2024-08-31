package parse

import (
	"strconv"
	"strings"
)

func getLineWithNumberAndContent(lineIndex int, lineContents string, addCaret bool) string {
	line := strconv.Itoa(lineIndex)

	if lineContents != "" {
		if addCaret {
			line = line + " > "
		} else {
			line = line + "   "
		}

		line = line + lineContents
	}

	return line
}

func (s *sectionedTemplate) setFatalMessage(msg string, expandedSourceLineIndex int) {
	var templateContext []string

	expandedTemplateLine := s.expandedTemplateLines[expandedSourceLineIndex]

	errorLine := expandedTemplateLine.sourceLineIndex
	lineBefore := errorLine - 1
	if lineBefore >= 0 {
		templateContext = append(templateContext, getLineWithNumberAndContent(lineBefore+1, s.rawTemplateLines[lineBefore], false))
	}

	templateContext = append(templateContext, getLineWithNumberAndContent(errorLine+1, s.rawTemplateLines[errorLine], true))

	lineAfter := errorLine + 1
	if lineAfter < len(s.rawTemplateLines) {
		templateContext = append(templateContext, getLineWithNumberAndContent(lineAfter+1, s.rawTemplateLines[lineAfter], false))
	}

	message := msg + " on line " + strconv.Itoa(errorLine+1) + ":\n"
	message = message + strings.Join(templateContext, "\n")

	if expandedTemplateLine.expanded {
		expandedMsg := "\nExpanded context:"
		beforeLine, nextLine := expandedSourceLineIndex-1, expandedSourceLineIndex+1

		if beforeLine > -1 && s.expandedTemplateLines[beforeLine].expanded {
			expandedMsg = expandedMsg + "\n" + getLineWithNumberAndContent(s.expandedTemplateLines[beforeLine].sourceLineIndex+1, s.expandedTemplateLines[beforeLine].String(), false)
		}

		expandedMsg = expandedMsg + "\n" + getLineWithNumberAndContent(expandedTemplateLine.sourceLineIndex+1, expandedTemplateLine.String(), true)

		if nextLine < len(s.expandedTemplateLines) && s.expandedTemplateLines[nextLine].expanded {
			expandedMsg = expandedMsg + "\n" + getLineWithNumberAndContent(s.expandedTemplateLines[nextLine].sourceLineIndex+1, s.expandedTemplateLines[nextLine].String(), false)
		}

		message = message + expandedMsg
	}

	s.fatals = append(s.fatals, message)
}

func (s *sectionedTemplate) getFatalMessages() string {
	fatalMessage := "Fatal error"
	if len(s.fatals) > 1 {
		fatalMessage = fatalMessage + "s"
	}

	fatalMessage = fatalMessage + " in file: " + s.filename + "\n"

	return fatalMessage + strings.Join(s.fatals, "\n\n")
}

func (s *sectionedTemplate) hasFatalMessages() bool {
	return len(s.fatals) > 0
}
