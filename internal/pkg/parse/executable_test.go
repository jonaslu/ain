package parse

import (
	"reflect"
	"testing"
)

func Test_getBalancedParensExecutablesGoodCases(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected []string
	}{
		"empty line returns empty result": {
			"",
			[]string{},
		},
		"no ${} expression returns empty result": {
			"nope, nothing here to see",
			[]string{},
		},
		"gets all executables on a line": {
			"$(one) $(two) $(three)",
			[]string{"$(one)", "$(two)", "$(three)"},
		},
		"empty executable": {
			`$()`,
			[]string{`$()`},
		},
		`handles parens in quotes, escaped and mixed quotes`: {
			`$(cmd '(arg1 \' "" arg2)' arg3 "arg4" "(arg5 '')")`,
			[]string{`$(cmd '(arg1 \' "" arg2)' arg3 "arg4" "(arg5 '')")`},
		},
	}

	for name, test := range tests {
		res, fatal := getExecutableExpr(test.input)
		if fatal != "" {
			t.Fatalf("Test: %s, got unexpected fatal: %v", name, fatal)
		}

		if !reflect.DeepEqual(test.expected, res) {
			t.Fatalf("Test: %s, got: %v, expected: %v", name, res, test.expected)
		}
	}
}

func Test_getBalancedParensExecutablesBadCases(t *testing.T) {
	tests := map[string]struct {
		input string
		fatal string
	}{
		"Unterminated quote sequence": {
			`$(cmd '(what now))`,
			`Unterminated quote sequence: $(cmd '(wh...`,
		},
		"Unterminated parens sans quote": {
			`$(cmd (yes but no`,
			"Missing end parenthesis on executable: $(cm...",
		},
		"Unterminated parens when quoting": {
			`$(cmd '(yes but no))'`,
			"Missing end parenthesis on executable: $(cm...",
		},
	}

	for name, test := range tests {
		_, fatal := getExecutableExpr(test.input)

		if test.fatal != fatal {
			t.Fatalf("Test: %s, got fatal: %v, expected: %v", name, fatal, test.fatal)
		}
	}
}
