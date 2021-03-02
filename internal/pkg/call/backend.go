package call

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/jonaslu/ain/internal/pkg/utils"
	"github.com/pkg/errors"
)

// !! TODO !! Make this a global config
const backendTimeoutSeconds = 10

func getTemplate(backend string) string {
	switch backend {
	case "curl":
		return `curl -sS
		{{if .Method}} -X {{.Method | ToUpper }} {{end}}
		{{range .Headers}} -H "{{.}}" {{end}}
		{{.Host.String}}`

	case "httpie":
		return `http --ignore-stdin
		{{if .Method}} {{.Method | ToUpper }} {{end}}
		{{.Host.String}}
		{{range .Headers}} "{{.}}" {{end}}`
	}

	return ""
}

func getFuncMap() map[string]interface{} {
	return template.FuncMap{
		"ToUpper": strings.ToUpper,
	}
}

func CallBackend(ctx context.Context, callData *Data, backend string) (string, error) {
	backendTemplateStr := getTemplate(backend)
	if backendTemplateStr == "" {
		return "", errors.Errorf("Template for backend: %s not found", backend)
	}

	backendTemplate, err := template.New("backend").Funcs(getFuncMap()).Parse(backendTemplateStr)
	if err != nil {
		return "", errors.Wrap(err, "Could not parse template")
	}

	var templateOutputBuilder strings.Builder
	err = backendTemplate.Execute(&templateOutputBuilder, callData)
	if err != nil {
		return "", errors.Wrap(err, "Could not execute template with callData")
	}

	templateOutput := templateOutputBuilder.String()
	templateOutput = strings.TrimSpace(templateOutput)

	if templateOutput == "" {
		return "", errors.New("Empty backend template result")
	}

	tokenizedCommandLine, err := utils.TokenizeLine(templateOutput, true)
	if err != nil {
		return "", errors.Wrap(err, "Error tokenizing backend template")
	}

	command := tokenizedCommandLine[0]

	if command == "" {
		return "", errors.Errorf("Empty backend command. Template output: %s", templateOutput)
	}

	args := tokenizedCommandLine[1:]

	backendTimeoutContext, _ := context.WithTimeout(ctx, backendTimeoutSeconds*time.Second)

	var stdout, stderr bytes.Buffer
	backendCmd := exec.CommandContext(backendTimeoutContext, command, args...)
	backendCmd.Stdout = &stdout
	backendCmd.Stderr = &stderr

	err = backendCmd.Run()
	stdoutStr := string(stdout.Bytes())

	if err != nil {
		stderrStr := string(stderr.Bytes())
		return "", errors.Errorf("Error: %v, running backend command: %s.\nError output: %s %s", err, backendCmd.String(), stderrStr, stdoutStr)
	}

	return stdoutStr, nil
}
