package parse

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func Test_sectionedTemplate_expandEnvVars_GoodCases(t *testing.T) {
	tests := map[string]struct {
		beforeTest     func()
		inputTemplate  string
		expectedResult []expandedSourceMarker
	}{
		"Substitution works": {
			beforeTest: func() {
				os.Setenv("VAR1", "value1")
				os.Setenv("VAR2", "value2")
			},
			inputTemplate: "${VAR1} ${VAR2}",
			expectedResult: []expandedSourceMarker{{
				content:         "value1 value2",
				fatalContent:    "value1 value2",
				comment:         "",
				sourceLineIndex: 0,
				expanded:        true,
			}}},
		"Fatal context keeps quoted envvars": {
			beforeTest: func() {
				os.Setenv("VAR1", "value1")
			},
			inputTemplate: "${VAR1} `${VAR2}",
			expectedResult: []expandedSourceMarker{{
				content:         "value1 ${VAR2}",
				fatalContent:    "value1 `${VAR2}",
				comment:         "",
				sourceLineIndex: 0,
				expanded:        true,
			}},
		},
	}
	for name, test := range tests {
		test.beforeTest()
		s := newSectionedTemplate(test.inputTemplate, "")

		if s.substituteEnvVars(); s.hasFatalMessages() {
			t.Errorf("Got unexpected fatals, %s ", s.getFatalMessages())
		} else {
			if !reflect.DeepEqual(test.expectedResult, s.expandedTemplateLines) {
				t.Errorf("Test: %s. Expected %v, got: %v", name, test.expectedResult, s.expandedTemplateLines)
			}
		}
	}
}

func Test_sectionedTemplate_expandEnvVars_BadCases(t *testing.T) {
	tests := map[string]struct {
		beforeTest           func()
		input                string
		expectedFatalMessage string
	}{
		"Empty variable": {
			beforeTest:           func() {},
			input:                "${}",
			expectedFatalMessage: "Empty variable",
		},
		"Cannot find value for variable": {
			beforeTest: func() {
				os.Unsetenv("VAR")
			},
			input:                "${VAR}",
			expectedFatalMessage: "Cannot find value for variable VAR",
		},
		"Value for variable is empty": {
			beforeTest: func() {
				os.Setenv("VAR", "")
			},
			input:                "${VAR}",
			expectedFatalMessage: "Value for variable VAR is empty",
		},
	}

	for name, test := range tests {
		test.beforeTest()
		s := newSectionedTemplate(test.input, "")
		s.substituteEnvVars()

		if len(s.fatals) != 1 {
			t.Errorf("Test: %s. Wrong number of fatals", name)
		}

		if !strings.Contains(s.fatals[0], test.expectedFatalMessage) {
			t.Errorf("Test: %s. Unexpected error message: %s", name, s.fatals[0])
		}
	}
}
