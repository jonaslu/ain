package main

import (
	"regexp"

	"github.com/jonaslu/ain/internal/pkg/utils"
)

var executableExpressionRe = regexp.MustCompile(`(m?)\$\([^)]*\)?`)
var executableRe = regexp.MustCompile(`\$\(([^)]*)\)`)

var emptyOutputLineRe = regexp.MustCompile(`^\s*$`)

type executableAndArgs struct {
	executable string
	args       []string
}

var includedSections = []string{
	HostSection,
	QuerySection,
	HeadersSection,
	MethodSection,
	BodySection,
	BackendSection,
	BackendOptionsSection,
	DefaultVarsSection,
}

func (s *SectionedTemplate) captureExecutableAndArgs() []executableAndArgs {
	executables := []executableAndArgs{}

	for _, sectionName := range includedSections {
		for _, templateLine := range *s.GetNamedSection(sectionName) {
			lineContents := templateLine.LineContents

			for _, executableWithParens := range executableExpressionRe.FindAllString(lineContents, -1) {
				executableAndArgsCapture := executableRe.FindStringSubmatch(executableWithParens)

				if len(executableAndArgsCapture) != 2 {
					s.SetFatalMessage("Malformed executable", templateLine.SourceLineIndex)
					continue
				}

				executableAndArgsStr := executableAndArgsCapture[1]
				if executableAndArgsStr == "" {
					s.SetFatalMessage("Empty executable", templateLine.SourceLineIndex)
					continue
				}

				tokenizedExecutableLine, err := utils.TokenizeLine(executableAndArgsStr)
				if err != nil {
					s.SetFatalMessage(err.Error(), templateLine.SourceLineIndex)
					continue
				}

				executable := tokenizedExecutableLine[0]

				executables = append(executables, executableAndArgs{
					executable: executable,
					args:       tokenizedExecutableLine[1:],
				})
			}
		}
	}

	return executables
}
