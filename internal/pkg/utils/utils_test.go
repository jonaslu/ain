package utils

import (
	"reflect"
	"testing"
)

func TestTokenizeLineGoodCases(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected []string
	}{
		`discard spaces`: {
			`   word   `,
			[]string{`word`},
		},

		`collect last word in buffer`: {
			`word`,
			[]string{`word`},
		},

		`split on two words`: {
			`word1 word2`,
			[]string{`word1`, `word2`},
		},

		`escaped quote`: {
			`wo\"rd`,
			[]string{`wo"rd`},
		},

		`escaped quote when quoting`: {
			`"wo\"rd"`,
			[]string{`wo"rd`},
		},

		`other quote when quoting`: {
			`"wo\'rd"`,
			[]string{`wo\'rd`},
		},

		`quote at edge`: {
			`word\"`,
			[]string{`word"`},
		},

		`quoting in the middle`: {
			`wo"  rd  "`,
			[]string{`wo  rd  `},
		},

		`quote and quote back to back`: {
			`"word""word"`,
			[]string{`wordword`},
		},

		`quoted whitespace in the beginning and end should be retained`: {
			`"   word    "`,
			[]string{`   word    `},
		},
	}

	for name, test := range tests {
		res, err := TokenizeLine(test.input)
		if err != nil {
			t.Fatalf("Got unexpected error: %v", err)
		}

		if !reflect.DeepEqual(test.expected, res) {
			t.Fatalf("Test: %s, got: %v, expected: %v", name, res, test.expected)
		}
	}
}

func TestTokenizeLineBadCases(t *testing.T) {
	tests := map[string]struct {
		input        string
		errorMessage string
	}{
		`only quote`: {
			`"`,
			`Unterminated quote sequence: "`,
		},

		`beginning of word`: {
			`"word`,
			`Unterminated quote sequence: "wor...`,
		},

		`end of word`: {
			`word"`,
			`Unterminated quote sequence: ...ord"`,
		},

		`no ellipsize on three chars beginning`: {
			`"the`,
			`Unterminated quote sequence: "the`,
		},

		`no ellipsize on three chars end`: {
			`the'`,
			`Unterminated quote sequence: the'`,
		},

		`full context no ellipsize in middle of word`: {
			`the'eht`,
			`Unterminated quote sequence: the'eht`,
		},

		`full context in middle of word`: {
			`word'drow`,
			`Unterminated quote sequence: ...ord'dro...`,
		},
	}

	for name, test := range tests {
		_, err := TokenizeLine(test.input)
		if err == nil {
			t.Fatal("No expected error")
		}

		if err.Error() != test.errorMessage {
			t.Fatalf("Test: %s, got error message: %v, expected: %v", name, err.Error(), test.errorMessage)
		}
	}
}
