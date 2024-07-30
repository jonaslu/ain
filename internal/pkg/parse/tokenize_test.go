package parse

import (
	"reflect"
	"testing"
)

func Test_unescapeTextContent(t *testing.T) {
	tests := map[string]struct {
		input          string
		expectedResult string
	}{
		"Everything but no escaped ending": {
			input:          "`${} `$() `#",
			expectedResult: "${} $() #",
		},
		"Everything and escaped ending": {
			input:          "`${} `$() `# \\`",
			expectedResult: "${} $() # `",
		},
	}

	for name, test := range tests {
		result := unescapeTextContent(test.input, envVarToken, true)
		if !reflect.DeepEqual(test.expectedResult, result) {
			t.Errorf("Test: %s, Expected: %v, Got: %v", name, test.expectedResult, result)
		}
	}
}

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

func Test_TokenizeGoodCases(t *testing.T) {
	tests := map[string]struct {
		input          string
		allowedContent tokenType
		expectedResult []token
	}{
		"Empty string": {
			input:          "",
			allowedContent: envVarToken,
			expectedResult: []token{},
		},
		"Only text": {
			input:          "abc 123",
			allowedContent: envVarToken,
			expectedResult: []token{{
				tokenType:    textToken,
				content:      "abc 123",
				fatalContent: "abc 123",
			}},
		},
		"Only comment": {
			input:          "# This is a comment",
			allowedContent: envVarToken,
			expectedResult: []token{
				{
					tokenType:    commentToken,
					fatalContent: "# This is a comment",
				},
			},
		},
		"Only envvar": {
			input:          "${ENV_VAR}",
			allowedContent: envVarToken,
			expectedResult: []token{
				{
					tokenType:    envVarToken,
					content:      "ENV_VAR",
					fatalContent: "${ENV_VAR}",
				},
			},
		},
		"Only executable": {
			input:          "$(executable)",
			allowedContent: envVarToken,
			expectedResult: []token{
				{
					tokenType:    executableToken,
					content:      "executable",
					fatalContent: "$(executable)",
				},
			},
		},
		"Envvar at ExecutableContent level": {
			input:          "${ENV_VAR}",
			allowedContent: executableToken,
			expectedResult: []token{
				{
					tokenType:    textToken,
					content:      "${ENV_VAR}",
					fatalContent: "${ENV_VAR}",
				},
			},
		},
		"Text with comment": {
			input:          "abc 123 # comment ${envvar} $(executable)",
			allowedContent: envVarToken,
			expectedResult: []token{{
				tokenType:    textToken,
				content:      "abc 123 ",
				fatalContent: "abc 123 ",
			}, {
				tokenType:    commentToken,
				fatalContent: "# comment ${envvar} $(executable)",
			}},
		},
		"Envvar not parsed but executable parsed at ExecutableContet level": {
			input:          "This ${is} a $(test) # yak ${envvar} ${executable}",
			allowedContent: executableToken,
			expectedResult: []token{
				{
					tokenType:    textToken,
					content:      "This ${is} a ",
					fatalContent: "This ${is} a ",
				},
				{
					tokenType:    executableToken,
					content:      "test",
					fatalContent: "$(test)",
				},
				{
					tokenType:    textToken,
					content:      " ",
					fatalContent: " ",
				},
				{
					tokenType:    commentToken,
					fatalContent: "# yak ${envvar} ${executable}",
				},
			},
		},
		"Envvar and executable not parsed at TextContent level": {
			input:          "This ${is} a $(test) # yak ${envvar} ${executable}",
			allowedContent: textToken,
			expectedResult: []token{
				{
					tokenType:    textToken,
					content:      "This ${is} a $(test) ",
					fatalContent: "This ${is} a $(test) ",
				},
				{
					tokenType:    commentToken,
					fatalContent: "# yak ${envvar} ${executable}",
				},
			},
		},
		"Multiple envvars and executables": {
			input:          "This is ${envvar1}${envvar2} and $(executable1)$(executable2)# comment 1 # comment2",
			allowedContent: envVarToken,
			expectedResult: []token{
				{
					tokenType:    textToken,
					content:      "This is ",
					fatalContent: "This is ",
				},
				{
					tokenType:    envVarToken,
					content:      "envvar1",
					fatalContent: "${envvar1}",
				},
				{
					tokenType:    envVarToken,
					content:      "envvar2",
					fatalContent: "${envvar2}",
				},
				{
					tokenType:    textToken,
					content:      " and ",
					fatalContent: " and ",
				},
				{
					tokenType:    executableToken,
					content:      "executable1",
					fatalContent: "$(executable1)",
				},
				{
					tokenType:    executableToken,
					content:      "executable2",
					fatalContent: "$(executable2)",
				},
				{
					tokenType:    commentToken,
					fatalContent: "# comment 1 # comment2",
				},
			},
		},
		"Escaping of all types at EnvVarContent level": {
			input:          "`${escaped} `$(escaped) `# escaped",
			allowedContent: envVarToken,
			expectedResult: []token{
				{
					tokenType:    textToken,
					content:      "${escaped} $(escaped) # escaped",
					fatalContent: "`${escaped} `$(escaped) `# escaped",
				},
			},
		},
		"Escaping of executable and comment at ExecutableContent level": {
			input:          "${escaped} `$(escaped) `# escaped",
			allowedContent: executableToken,
			expectedResult: []token{
				{
					tokenType:    textToken,
					content:      "${escaped} $(escaped) # escaped",
					fatalContent: "${escaped} `$(escaped) `# escaped",
				},
			},
		},
		"Escaping of comment at TextContent level": {
			input:          "${not escaped} $(not escaped) `# escaped",
			allowedContent: textToken,
			expectedResult: []token{
				{
					tokenType:    textToken,
					content:      "${not escaped} $(not escaped) # escaped",
					fatalContent: "${not escaped} $(not escaped) `# escaped",
				},
			},
		},
		"Escaped end bracket inside envvar": {
			input:          "${yeti`}`}\\`}",
			allowedContent: envVarToken,
			expectedResult: []token{{
				tokenType:    envVarToken,
				content:      "yeti}}`",
				fatalContent: "${yeti`}`}\\`}",
			}},
		},
		"Escaped end parenthesis inside executable": {
			input:          "$(echo `)yo`)\\`)",
			allowedContent: envVarToken,
			expectedResult: []token{{
				tokenType:    executableToken,
				content:      "echo )yo)`",
				fatalContent: "$(echo `)yo`)\\`)",
			}},
		},
		"Executable no need to escape ) when quoting": {
			input:          "$(echo \")\\\")\"`) '))')",
			allowedContent: envVarToken,
			expectedResult: []token{{
				tokenType:    executableToken,
				content:      "echo \")\\\")\") '))'",
				fatalContent: "$(echo \")\\\")\"`) '))')",
			}},
		},
	}

	for name, test := range tests {
		result, _ := Tokenize(test.input, test.allowedContent)

		if !reflect.DeepEqual(test.expectedResult, result) {
			t.Errorf("Test: %s, Allowed Content: %d, Expected: %v, Got: %v", name, test.allowedContent, test.expectedResult, result)
		}
	}
}

func Test_TokenizeBadCases(t *testing.T) {
	tests := map[string]struct {
		input          string
		allowedContent tokenType
		expectedFatal  string
	}{
		"Missing bracket for empty envvar": {
			input:          "${",
			allowedContent: envVarToken,
			expectedFatal:  "Missing closing bracket for environment variable: ${",
		},
		"Missing parenthesis for empty executable": {
			input:          "$(",
			allowedContent: executableToken,
			expectedFatal:  "Missing closing parenthesis for executable: $(",
		},
		"Missing bracket for envvar with content": {
			input:          "${envvar `}",
			allowedContent: envVarToken,
			expectedFatal:  "Missing closing bracket for environment variable: ${envvar `}",
		},
		"Missing parenthesis for executable with content": {
			input:          "$(executable arg1 arg2 `)",
			allowedContent: executableToken,
			expectedFatal:  "Missing closing parenthesis for executable: $(executable arg1 arg2 `)",
		},
		`Missing end " quote for executable`: {
			input:          `$(executable ")`,
			allowedContent: envVarToken,
			expectedFatal:  `Unterminated quote sequence for executable: $(executable ")`,
		},
		`Missing end ' quote for executable`: {
			input:          `$(executable ')`,
			allowedContent: envVarToken,
			expectedFatal:  `Unterminated quote sequence for executable: $(executable ')`,
		},
	}

	for name, test := range tests {
		_, fatal := Tokenize(test.input, test.allowedContent)

		if fatal != test.expectedFatal {
			t.Errorf("Test: %s, Expected: %v, Got: %v", name, test.expectedFatal, fatal)
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
		input             string
		expectedTokens    []token
		expectedHasTokens bool
	}{
		"Empty input": {
			input:             "",
			expectedTokens:    []token{},
			expectedHasTokens: false,
		},
		"Only text": {
			input: "teeext",
			expectedTokens: []token{{
				tokenType:    textToken,
				content:      "teeext",
				fatalContent: "teeext",
			}},
			expectedHasTokens: false,
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
			expectedHasTokens: true,
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
			expectedHasTokens: true,
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

			expectedHasTokens: true,
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
			expectedHasTokens: true,
		},
		"Escaped end bracket in envvar": {
			input: "${VAR1`}}",
			expectedTokens: []token{{
				tokenType:    envVarToken,
				content:      "VAR1}",
				fatalContent: "${VAR1`}}",
			}},
			expectedHasTokens: true,
		},
		"Escaped backtick last in envvar": {
			input: "${ENV\\`}",
			expectedTokens: []token{{
				tokenType:    envVarToken,
				content:      "ENV`",
				fatalContent: "${ENV\\`}",
			}},
			expectedHasTokens: true,
		},
	}

	for name, test := range tests {
		tokens, hasEnvVarTokens, fatal := tokenizeEnvVars(test.input)
		if fatal != "" {
			t.Errorf("Test: %s, Unexpected fatal: %s", name, fatal)
		}

		if test.expectedHasTokens != hasEnvVarTokens {
			t.Errorf("Test: %s, Expected hasEnvVarTokens: %v, Got: %v", name, test.expectedHasTokens, hasEnvVarTokens)
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
		_, _, fatal := tokenizeEnvVars(test.input)
		if fatal != test.expectedFatal {
			t.Errorf("Test: %s, Expected fatal: %v, Got: %v", name, test.expectedFatal, fatal)
		}
	}
}

func Test_tokenizeExecutablesGoodCases(t *testing.T) {
	tests := map[string]struct {
		input             string
		expectedTokens    []token
		expectedHasTokens bool
	}{
		"Empty input": {
			input:             "",
			expectedTokens:    []token{},
			expectedHasTokens: false,
		},
		"Only text": {
			input: "teeext",
			expectedTokens: []token{{
				tokenType:    textToken,
				content:      "teeext",
				fatalContent: "teeext",
			}},
			expectedHasTokens: false,
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
			expectedHasTokens: true,
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
			expectedHasTokens: true,
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
			expectedHasTokens: true,
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
			expectedHasTokens: true,
		},
		"Escaped end parenthesis inside executable": {
			input: "$(echo `)yo`)\\`)",
			expectedTokens: []token{{
				tokenType:    executableToken,
				content:      "echo )yo)`",
				fatalContent: "$(echo `)yo`)\\`)",
			}},
			expectedHasTokens: true,
		},
		"Executable no need to escape ) when quoting": {
			input: "$(echo \")\\\")\"`) '))')",
			expectedTokens: []token{{
				tokenType:    executableToken,
				content:      "echo \")\\\")\") '))'",
				fatalContent: "$(echo \")\\\")\"`) '))')",
			}},
			expectedHasTokens: true,
		},
	}

	for name, test := range tests {
		tokens, hasExecutableTokens, fatal := tokenizeExecutables(test.input)
		if fatal != "" {
			t.Errorf("Test: %s, Unexpected fatal: %s", name, fatal)
		}

		if test.expectedHasTokens != hasExecutableTokens {
			t.Errorf("Test: %s, Expected hasEnvVarTokens: %v, Got: %v", name, test.expectedHasTokens, hasExecutableTokens)
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
		_, _, fatal := tokenizeExecutables(test.input)
		if fatal != test.expectedFatal {
			t.Errorf("Test: %s, Expected fatal: %v, Got: %v", name, test.expectedFatal, fatal)
		}
	}
}
