package parse

import (
	"reflect"
	"strings"
	"testing"
)

func Test_sectionedTemplate_insertExecutableOutputGoodCases(t *testing.T) {
	tests := map[string]struct {
		inputTemplate     string
		executableResults *[]executableOutput
		expectedResult    []expandedSourceMarker
	}{
		"Executable output inserted": {
			inputTemplate: "$(cmd)",
			executableResults: &[]executableOutput{{
				cmdOutput:    "cmd output",
				fatalMessage: "",
			}},
			expectedResult: []expandedSourceMarker{{
				content:         "cmd output",
				fatalContent:    "cmd output",
				comment:         "",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
		"Escaped quoting kept in fatal context": {
			inputTemplate: "$(cmd1) `$(cmd2)",
			executableResults: &[]executableOutput{{
				cmdOutput:    "cmd1 output",
				fatalMessage: "",
			}},
			expectedResult: []expandedSourceMarker{{
				content:         "cmd1 output $(cmd2)",
				fatalContent:    "cmd1 output `$(cmd2)",
				comment:         "",
				sourceLineIndex: 0,
				expanded:        true,
			}},
		},
	}
	for name, test := range tests {
		s := newSectionedTemplate(test.inputTemplate, "")

		if s.insertExecutableOutput(test.executableResults); s.hasFatalMessages() {
			t.Errorf("Got unexpected fatals, %s ", s.getFatalMessages())
		} else {
			if !reflect.DeepEqual(test.expectedResult, s.expandedTemplateLines) {
				t.Errorf("Test: %s. Expected %v, got: %v", name, test.expectedResult, s.expandedTemplateLines)
			}
		}
	}
}

func Test_sectionedTemplate_insertExecutableOutputBadCases(t *testing.T) {
	tests := map[string]struct {
		inputTemplate        string
		executableResults    *[]executableOutput
		expectedFatalMessage string
	}{
		"Executable fatal returned": {
			inputTemplate: "$(cmd)",
			executableResults: &[]executableOutput{{
				cmdOutput:    "",
				fatalMessage: "This is the fatal message",
			}},
			expectedFatalMessage: "This is the fatal message",
		},
	}
	for name, test := range tests {
		s := newSectionedTemplate(test.inputTemplate, "")
		s.insertExecutableOutput(test.executableResults)

		if len(s.fatals) != 1 {
			t.Errorf("Test: %s. Wrong number of fatals", name)
		}

		if !strings.Contains(s.fatals[0], test.expectedFatalMessage) {
			t.Errorf("Test: %s. Unexpected error message: %s", name, s.fatals[0])
		}
	}
}
