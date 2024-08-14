package parse

import (
	"reflect"
	"strings"
	"testing"
)

func Test_sectionedTemplate_getTextContent(t *testing.T) {
	tests := map[string]struct {
		content        string
		comment        string
		expectedResult string
	}{
		"Escaped comments unescaped": {
			content:        "text `# no comment",
			comment:        "",
			expectedResult: "text # no comment",
		},
		"Quote unescaped if comment on line": {
			content:        "text \\`",
			comment:        "# comment",
			expectedResult: "text `",
		},
		"Quote left untouched if no comment": {
			content:        "text \\`",
			comment:        "",
			expectedResult: "text \\`",
		},
	}

	for name, test := range tests {
		s := expandedSourceMarker{
			content: test.content,
			comment: test.comment,
		}

		result := s.getTextContent()
		if test.expectedResult != result {
			t.Errorf("Test: %s. Expected %v, got: %v", name, result, test.expectedResult)
		}
	}
}

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
				content:         "VAR text ",
				fatalContent:    "VAR text ",
				comment:         "# comment",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
		"Double envvar substitution before comment": {
			inputTemplate: "${VAR1} ${VAR2} # comment",
			expectedResult: []expandedSourceMarker{{
				content:         "VAR1 VAR2 ",
				fatalContent:    "VAR1 VAR2 ",
				comment:         "# comment",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
		"Single envvar substitution with comment disables rest of line": {
			inputTemplate: "${VAR1 🐐 comment1} ${VAR2} # comment2",
			expectedResult: []expandedSourceMarker{{
				content:         "VAR1 ",
				fatalContent:    "VAR1 ",
				comment:         "# comment1 ${VAR2} # comment2",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
		"Single envvar with newline pushes rest of line one row below": {
			inputTemplate: "${VAR1🐷} ${VAR2} # comment",
			expectedResult: []expandedSourceMarker{{
				content:         "VAR1",
				fatalContent:    "VAR1",
				comment:         "",
				sourceLineIndex: 0,
				expanded:        true,
			}, {
				content:         " VAR2 ",
				fatalContent:    " VAR2 ",
				comment:         "# comment",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
		"Single envvar with newline and comment pushes rest of line one row below": {
			inputTemplate: "${VAR1 🐐 comment1🐷} ${VAR2} # comment2",
			expectedResult: []expandedSourceMarker{{
				content:         "VAR1 ",
				fatalContent:    "VAR1 ",
				comment:         "# comment1",
				sourceLineIndex: 0,
				expanded:        true,
			}, {
				content:         " VAR2 ",
				fatalContent:    " VAR2 ",
				comment:         "# comment2",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
		"Single envvar with newline, comment, newline and comment pushes rest of line one row below and disables": {
			inputTemplate: "${VAR1 🐐 comment1🐷🐐} ${VAR2} # comment2",
			expectedResult: []expandedSourceMarker{{
				content:         "VAR1 ",
				fatalContent:    "VAR1 ",
				comment:         "# comment1",
				sourceLineIndex: 0,
				expanded:        true,
			}, {
				content:         "",
				fatalContent:    "",
				comment:         "# ${VAR2} # comment2",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
	}

	for name, test := range tests {
		s := newSectionedTemplate(test.inputTemplate, "")

		if s.expandTemplateLines(tokenizeEnvVars, echoIterator); s.hasFatalMessages() {
			t.Errorf("Test: %s. Got unexpected fatals, %s ", name, s.getFatalMessages())
		} else {
			if !reflect.DeepEqual(test.expectedResult, s.expandedTemplateLines) {
				t.Errorf("Test: %s. Expected %v, got: %v", name, test.expectedResult, s.expandedTemplateLines)
			}
		}
	}
}
