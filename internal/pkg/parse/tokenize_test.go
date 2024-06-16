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
