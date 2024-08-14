package parse

import (
	"reflect"
	"testing"
)

func Test_isStartOfContentType(t *testing.T) {
	tests := map[string]struct {
		prefix, prev, rest string
		expectedResult     bool
	}{
		"Regular comment": {
			prefix:         "#",
			prev:           "goat",
			rest:           "# yak",
			expectedResult: true,
		},
		"Escaped comment": {
			prefix:         "#",
			prev:           "fire `",
			rest:           "# no comment",
			expectedResult: false,
		},
		"Escaped backtick followed by comment": {
			prefix:         "#",
			prev:           "\\`",
			rest:           "# is comment",
			expectedResult: true,
		},
	}

	for name, test := range tests {
		result := isStartOfToken(test.prefix, test.prev, test.rest)
		if !reflect.DeepEqual(test.expectedResult, result) {
			t.Errorf("Test: %s, Expected: %v, Got: %v", name, test.expectedResult, result)
		}
	}
}

func Test_splitTextOnComment(t *testing.T) {
	tests := map[string]struct {
		input           string
		expectedContent string
		expectedComment string
	}{
		"Only content": {
			input:           "abc 123",
			expectedContent: "abc 123",
			expectedComment: "",
		},
		"Only comment": {
			input:           "# uh comment yo",
			expectedContent: "",
			expectedComment: "# uh comment yo",
		},
		"Mixed content and comment": {
			input:           "abc123 # uh comment yo # more",
			expectedContent: "abc123 ",
			expectedComment: "# uh comment yo # more",
		},
		"Escaped comment": {
			input:           "abc123 `# still content",
			expectedContent: "abc123 `# still content",
			expectedComment: "",
		},
		"Backtick followed by comment": {
			input:           "abc123 \\`# comment",
			expectedContent: "abc123 \\`",
			expectedComment: "# comment",
		},
	}

	for name, test := range tests {
		content, comment := splitTextOnComment(test.input)
		if content != test.expectedContent {
			t.Errorf("Test: %s, Expected content: %v, Got: %v", name, test.expectedContent, content)
		}

		if comment != test.expectedComment {
			t.Errorf("Test: %s, Expected comment: %v, Got: %v", name, test.expectedComment, comment)
		}
	}
}

func Test_tokenizeEnvVarsGoodCases(t *testing.T) {
	tests := map[string]struct {
		input          string
		expectedTokens []token
	}{
		"Empty input": {
			input:          "",
			expectedTokens: []token{},
		},
		"Only text": {
			input: "teeext",
			expectedTokens: []token{{
				tokenType:    textToken,
				content:      "teeext",
				fatalContent: "teeext",
			}},
		},
		"Only envvars": {
			input: "${VAR1}${VAR2}",
			expectedTokens: []token{
				{
					tokenType:    envVarToken,
					content:      "VAR1",
					fatalContent: "${VAR1}",
				},
				{
					tokenType:    envVarToken,
					content:      "VAR2",
					fatalContent: "${VAR2}",
				},
			},
		},
		"Text and envvars (comments are not handled)": {
			input: "Ugh ${VAR1}",
			expectedTokens: []token{{
				tokenType:    textToken,
				content:      "Ugh ",
				fatalContent: "Ugh ",
			}, {
				tokenType:    envVarToken,
				content:      "VAR1",
				fatalContent: "${VAR1}",
			}},
		},
		"Escaped envvars converted to text": {
			input: "`${VAR1}\\`${VAR2}",
			expectedTokens: []token{
				{
					tokenType:    textToken,
					content:      "${VAR1}`",
					fatalContent: "`${VAR1}\\`",
				},
				{
					tokenType:    envVarToken,
					content:      "VAR2",
					fatalContent: "${VAR2}",
				},
			},
		},
		"Escaped backtick is literal at end of input": {
			input: "${VAR1}\\`",
			expectedTokens: []token{
				{
					tokenType:    envVarToken,
					content:      "VAR1",
					fatalContent: "${VAR1}",
				},
				{
					tokenType:    textToken,
					content:      "\\`",
					fatalContent: "\\`",
				},
			},
		},
		"Escaped end bracket in envvar": {
			input: "${VAR1`}}",
			expectedTokens: []token{{
				tokenType:    envVarToken,
				content:      "VAR1}",
				fatalContent: "${VAR1`}}",
			}},
		},
		"Escaped backtick last in envvar": {
			input: "${ENV\\`}",
			expectedTokens: []token{{
				tokenType:    envVarToken,
				content:      "ENV`",
				fatalContent: "${ENV\\`}",
			}},
		},
	}

	for name, test := range tests {
		tokens, fatal := tokenizeEnvVars(test.input)
		if fatal != "" {
			t.Errorf("Test: %s, Unexpected fatal: %s", name, fatal)
		}

		if !reflect.DeepEqual(test.expectedTokens, tokens) {
			t.Errorf("Test: %s, Expected tokens: %v, Got: %v", name, test.expectedTokens, tokens)
		}
	}
}

func Test_tokenizeEnvVarsBadCases(t *testing.T) {
	tests := map[string]struct {
		input         string
		expectedFatal string
	}{
		"Missing closing bracket for envvar": {
			input:         "${VAR ${VAR",
			expectedFatal: "Missing closing bracket for environment variable: ${VAR ${VAR",
		},
	}

	for name, test := range tests {
		_, fatal := tokenizeEnvVars(test.input)
		if fatal != test.expectedFatal {
			t.Errorf("Test: %s, Expected fatal: %v, Got: %v", name, test.expectedFatal, fatal)
		}
	}
}

func Test_tokenizeExecutablesGoodCases(t *testing.T) {
	tests := map[string]struct {
		input          string
		expectedTokens []token
	}{
		"Empty input": {
			input:          "",
			expectedTokens: []token{},
		},
		"Only text": {
			input: "teeext",
			expectedTokens: []token{{
				tokenType:    textToken,
				content:      "teeext",
				fatalContent: "teeext",
			}},
		},
		"Only executables": {
			input: "$(cmd1 arg1 arg2)$(cmd2 arg3 arg4)",
			expectedTokens: []token{{
				tokenType:    executableToken,
				content:      "cmd1 arg1 arg2",
				fatalContent: "$(cmd1 arg1 arg2)",
			}, {
				tokenType:    executableToken,
				content:      "cmd2 arg3 arg4",
				fatalContent: "$(cmd2 arg3 arg4)",
			}},
		},
		"Text and executables (comments are not handled)": {
			input: "text $(cmd1 arg1 arg2)",
			expectedTokens: []token{
				{
					tokenType:    textToken,
					content:      "text ",
					fatalContent: "text ",
				},
				{
					tokenType:    executableToken,
					content:      "cmd1 arg1 arg2",
					fatalContent: "$(cmd1 arg1 arg2)",
				},
			},
		},
		"Escaped executables converted to text": {
			input: "`$(cmd1)\\`$(cmd2)",
			expectedTokens: []token{{
				tokenType:    textToken,
				content:      "$(cmd1)`",
				fatalContent: "`$(cmd1)\\`",
			}, {
				tokenType:    executableToken,
				content:      "cmd2",
				fatalContent: "$(cmd2)",
			}},
		},
		"Escaped backtick is literal at end of input": {
			input: "$(cmd1)\\`",
			expectedTokens: []token{{
				tokenType:    executableToken,
				content:      "cmd1",
				fatalContent: "$(cmd1)",
			}, {
				tokenType:    textToken,
				content:      "\\`",
				fatalContent: "\\`",
			}},
		},
		"Escaped end parenthesis inside executable": {
			input: "$(echo `)yo`)\\`)",
			expectedTokens: []token{{
				tokenType:    executableToken,
				content:      "echo )yo)`",
				fatalContent: "$(echo `)yo`)\\`)",
			}},
		},
		"Executable no need to escape ) when quoting": {
			input: "$(echo \")\\\")\"`) '))')",
			expectedTokens: []token{{
				tokenType:    executableToken,
				content:      "echo \")\\\")\") '))'",
				fatalContent: "$(echo \")\\\")\"`) '))')",
			}},
		},
	}

	for name, test := range tests {
		tokens, fatal := tokenizeExecutables(test.input)
		if fatal != "" {
			t.Errorf("Test: %s, Unexpected fatal: %s", name, fatal)
		}

		if !reflect.DeepEqual(test.expectedTokens, tokens) {
			t.Errorf("Test: %s, Expected tokens: %v, Got: %v", name, test.expectedTokens, tokens)
		}
	}
}

func Test_tokenizeExecutablesBadCases(t *testing.T) {
	tests := map[string]struct {
		input         string
		expectedFatal string
	}{
		"Missing closing parenthesis for executable": {
			input:         "$(cmd1 $(cmd2",
			expectedFatal: "Missing closing parenthesis for executable: $(cmd1 $(cmd2",
		},
		"Unterminated quote sequence single quote": {
			input:         "$(node -e 'console.log(\"Yo yo\"))",
			expectedFatal: "Unterminated quote sequence for executable: $(node -e 'console.log(\"Yo yo\"))",
		},
		"Unterminated quote sequence double quote": {
			input:         "$(node -e \"console.log('Yo yo'))",
			expectedFatal: "Unterminated quote sequence for executable: $(node -e \"console.log('Yo yo'))",
		},
	}

	for name, test := range tests {
		_, fatal := tokenizeExecutables(test.input)
		if fatal != test.expectedFatal {
			t.Errorf("Test: %s, Expected fatal: %v, Got: %v", name, test.expectedFatal, fatal)
		}
	}
}
