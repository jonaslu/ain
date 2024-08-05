package parse

import (
	"reflect"
	"strings"
	"testing"
)

func Test_sectionedTemplate_expandTemplateLinesGoodCases(t *testing.T) {
	// Converts 🐐 to a comment (#)
	// Converts 🐷 to a newline
	echoIterator := func(c token) (string, string) {
		c.content = strings.ReplaceAll(c.content, "🐐", "#")
		c.content = strings.ReplaceAll(c.content, "🐷", "\n")

		return c.content, ""
	}

	tests := map[string]struct {
		inputTemplate  string
		expectedResult []expandedSourceMarker
	}{
		"Simple envvar substitution before comment": {
			inputTemplate: "${VAR} text # comment",
			expectedResult: []expandedSourceMarker{{
				tokens:          nil,
				content:         "VAR text ",
				comment:         "# comment",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
		"Double envvar substitution before comment": {
			inputTemplate: "${VAR1} ${VAR2} # comment",
			expectedResult: []expandedSourceMarker{{
				tokens:          nil,
				content:         "VAR1 VAR2 ",
				comment:         "# comment",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
		"Single envvar substitution with comment disables rest of line": {
			inputTemplate: "${VAR1 🐐 comment1} ${VAR2} # comment2",
			expectedResult: []expandedSourceMarker{{
				tokens:          nil,
				content:         "VAR1 ",
				comment:         "# comment1 ${VAR2} # comment2",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
		"Single envvar with newline pushes rest of line one row below": {
			inputTemplate: "${VAR1🐷} ${VAR2} # comment",
			expectedResult: []expandedSourceMarker{{
				tokens:          nil,
				content:         "VAR1",
				comment:         "",
				sourceLineIndex: 0,
				expanded:        true,
			}, {
				tokens:          nil,
				content:         " VAR2 ",
				comment:         "# comment",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
		"Single envvar with newline and comment pushes rest of line one row below": {
			inputTemplate: "${VAR1 🐐 comment1🐷} ${VAR2} # comment2",
			expectedResult: []expandedSourceMarker{{
				tokens:          nil,
				content:         "VAR1 ",
				comment:         "# comment1",
				sourceLineIndex: 0,
				expanded:        true,
			}, {
				tokens:          nil,
				content:         " VAR2 ",
				comment:         "# comment2",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
		"Single envvar with newline, comment, newline and comment pushes rest of line one row below and disables": {
			inputTemplate: "${VAR1 🐐 comment1🐷🐐} ${VAR2} # comment2",
			expectedResult: []expandedSourceMarker{{
				tokens:          nil,
				content:         "VAR1 ",
				comment:         "# comment1",
				sourceLineIndex: 0,
				expanded:        true,
			}, {
				tokens:          nil,
				content:         "",
				comment:         "# ${VAR2} # comment2",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
	}

	for name, test := range tests {
		s := newSectionedTemplate2(test.inputTemplate, "")

		if s.expandTemplateLines(tokenizeEnvVars, echoIterator); s.hasFatalMessages() {
			t.Errorf("Got unexpected fatals, %s ", s.getFatalMessages())
		} else {
			if !reflect.DeepEqual(test.expectedResult, s.expandedTemplateLines) {
				t.Errorf("Test: %s. Expected %v, got: %v", name, test.expectedResult, s.expandedTemplateLines)
			}
		}
	}
}
