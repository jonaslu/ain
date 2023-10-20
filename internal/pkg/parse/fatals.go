package parse

import (
	"strconv"
	"strings"
)

func (s *sectionedTemplate) setFatalMessage(msg string, sourceLineIndex int) {
	var templateContext []string

	errorLine := sourceLineIndex
	lineBefore := errorLine - 1
	if lineBefore >= 0 {
		templateContext = append(templateContext, strconv.Itoa(lineBefore+1)+"   "+s.rawTemplateLines[lineBefore])
	}

	templateContext = append(templateContext, strconv.Itoa(errorLine+1)+" > "+s.rawTemplateLines[errorLine])

	lineAfter := errorLine + 1
	if lineAfter < len(s.rawTemplateLines) {
		templateContext = append(templateContext, strconv.Itoa(lineAfter+1)+"   "+s.rawTemplateLines[lineAfter])
	}

	message := msg + " on line " + strconv.Itoa(sourceLineIndex+1) + ":\n"
	message = message + strings.Join(templateContext, "\n")

	s.fatals = append(s.fatals, message)
}

func (s *sectionedTemplate) getFatalMessages() string {
	fatalMessage := "Fatal error"
	if len(s.fatals) > 1 {
		fatalMessage = fatalMessage + "s"
	}

	fatalMessage = fatalMessage + " in file: " + s.filename + "\n"

	return fatalMessage + strings.Join(s.fatals, "\n")
}

func (s *sectionedTemplate) hasFatalMessages() bool {
	return len(s.fatals) > 0
}
