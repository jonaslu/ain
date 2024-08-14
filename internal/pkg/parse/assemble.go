package parse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/disk"
)

func getAllSectionedTemplates(filenames []string) ([]*sectionedTemplate, []string, error) {
	newSectionedTemplateFatals := []string{}
	allSectionedTemplates := []*sectionedTemplate{}

	for _, filename := range filenames {
		// !! TODO !! The file-name will be displayed as test.ain! <- Remove the exclamation-mark
		// when setting the file-name2.
		rawTemplateString, err := disk.ReadRawTemplateString(filename)
		if err != nil {
			return nil, newSectionedTemplateFatals, err
		}

		// !! TODO !! newSectionedTemplate does not set fatals anymore
		if sectionedTemplate := newSectionedTemplate(rawTemplateString, filename); sectionedTemplate.hasFatalMessages() {
			newSectionedTemplateFatals = append(newSectionedTemplateFatals, sectionedTemplate.getFatalMessages())
		} else {
			allSectionedTemplates = append(allSectionedTemplates, sectionedTemplate)
		}
	}

	return allSectionedTemplates, newSectionedTemplateFatals, nil
}

func getConfig(allSectionedTemplates []*sectionedTemplate) (data.Config, []string) {
	configFatals := []string{}
	config := data.NewConfig()

	for i := len(allSectionedTemplates) - 1; i >= 0; i-- {
		sectionedTemplate := allSectionedTemplates[i]

		if sectionedTemplate.setCapturedSections(configSection); sectionedTemplate.hasFatalMessages() {
			configFatals = append(configFatals, sectionedTemplate.getFatalMessages())
			break
		}

		localConfig := sectionedTemplate.getConfig()
		if sectionedTemplate.hasFatalMessages() {
			configFatals = append(configFatals, sectionedTemplate.getFatalMessages())
			break
		}

		if config.Timeout == data.TimeoutNotSet {
			config.Timeout = localConfig.Timeout
		}

		if config.QueryDelim == nil {
			config.QueryDelim = localConfig.QueryDelim
		}

		if config.Timeout > data.TimeoutNotSet && config.QueryDelim != nil {
			break
		}
	}

	return config, configFatals
}

func substituteEnvVars(allSectionedTemplates []*sectionedTemplate) []string {
	substituteEnvVarsFatals := []string{}

	for _, sectionedTemplate := range allSectionedTemplates {
		if sectionedTemplate.substituteEnvVars(); sectionedTemplate.hasFatalMessages() {
			substituteEnvVarsFatals = append(substituteEnvVarsFatals, sectionedTemplate.getFatalMessages())
		}
	}

	return substituteEnvVarsFatals
}

func substituteExecutables(ctx context.Context, config data.Config, allSectionedTemplates []*sectionedTemplate) ([]string, error) {
	substituteExecutablesFatals := []string{}
	allExecutableAndArgs := []executableAndArgs{}

	for _, sectionedTemplate := range allSectionedTemplates {
		allExecutableAndArgs = append(allExecutableAndArgs, sectionedTemplate.captureExecutableAndArgs()...)

		if sectionedTemplate.hasFatalMessages() {
			substituteExecutablesFatals = append(substituteExecutablesFatals, sectionedTemplate.getFatalMessages())
		}
	}

	if len(substituteExecutablesFatals) > 0 {
		return substituteExecutablesFatals, nil
	}

	allExecutablesOutput := callExecutables(ctx, config, allExecutableAndArgs)
	if ctx.Err() == context.Canceled {
		return nil, ctx.Err()
	}

	for _, sectionedTemplate := range allSectionedTemplates {
		if sectionedTemplate.insertExecutableOutput(&allExecutablesOutput); sectionedTemplate.hasFatalMessages() {
			substituteExecutablesFatals = append(substituteExecutablesFatals, sectionedTemplate.getFatalMessages())
		}
	}

	return substituteExecutablesFatals, nil
}

type allSectionRows struct {
	host           string
	backend        string
	method         string
	headers        []string
	query          []string
	body           []string
	backendOptions [][]string
}

func getAllSectionRows(allSectionedTemplates []*sectionedTemplate) (allSectionRows, []string) {
	allSectionRowsFatals := []string{}
	allSectionRows := allSectionRows{}

	for _, sectionedTemplate := range allSectionedTemplates {
		if sectionedTemplate.setCapturedSections(sectionsAllowingExecutables...); sectionedTemplate.hasFatalMessages() {
			allSectionRowsFatals = append(allSectionRowsFatals, sectionedTemplate.getFatalMessages())
			continue
		}

		allSectionRows.host = allSectionRows.host + sectionedTemplate.getHost()
		allSectionRows.headers = append(allSectionRows.headers, sectionedTemplate.getHeaders()...)
		allSectionRows.query = append(allSectionRows.query, sectionedTemplate.getQuery()...)
		allSectionRows.backendOptions = append(allSectionRows.backendOptions, sectionedTemplate.getBackendOptions()...)

		if localBackend := sectionedTemplate.getBackend(); localBackend != "" {
			allSectionRows.backend = localBackend
		}

		if localMethod := sectionedTemplate.getMethod(); localMethod != "" {
			allSectionRows.method = localMethod
		}

		if localBody := sectionedTemplate.getBody(); len(localBody) > 0 {
			allSectionRows.body = localBody
		}

		if sectionedTemplate.hasFatalMessages() {
			allSectionRowsFatals = append(allSectionRowsFatals, sectionedTemplate.getFatalMessages())
		}
	}

	return allSectionRows, allSectionRowsFatals
}

func getBackendInput(allSectionRows allSectionRows, config data.Config) (*data.BackendInput, []string) {
	backendInputFatals := []string{}
	backendInput := data.BackendInput{}

	if allSectionRows.host == "" {
		backendInputFatals = append(backendInputFatals, "No mandatory [Host] section found")
	} else {
		hostUrl, err := url.Parse(allSectionRows.host)

		if err != nil {
			backendInputFatals = append(backendInputFatals, fmt.Sprintf("[Host] has illegal url: %s, error: %v", allSectionRows.host, err))
		} else {
			addQueryString(hostUrl, allSectionRows.query, config)
			backendInput.Host = hostUrl
		}
	}

	if allSectionRows.backend == "" {
		backendInputFatals = append(backendInputFatals, "No mandatory [Backend] section found")
	}

	backendInput.Method = allSectionRows.method
	backendInput.Body = allSectionRows.body
	backendInput.Headers = allSectionRows.headers
	backendInput.Backend = allSectionRows.backend
	backendInput.BackendOptions = allSectionRows.backendOptions

	return &backendInput, backendInputFatals
}

func Assemble(ctx context.Context, filenames []string) (context.Context, *data.BackendInput, string, error) {
	allSectionedTemplates, allSectionedTemplateFatals, err := getAllSectionedTemplates(filenames)
	if err != nil {
		return ctx, nil, "", err
	}

	if len(allSectionedTemplateFatals) > 0 {
		return ctx, nil, strings.Join(allSectionedTemplateFatals, "\n\n"), nil
	}

	if substituteEnvVarsFatals := substituteEnvVars(allSectionedTemplates); len(substituteEnvVarsFatals) > 0 {
		return ctx, nil, strings.Join(substituteEnvVarsFatals, "\n\n"), nil
	}

	config, configFatals := getConfig(allSectionedTemplates)
	if len(configFatals) > 0 {
		return ctx, nil, strings.Join(configFatals, "\n\n"), nil
	}

	if config.Timeout != data.TimeoutNotSet {
		ctx, _ = context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
		ctx = context.WithValue(ctx, data.TimeoutContextValueKey{}, config.Timeout)
	}

	substituteExecutablesFatals, err := substituteExecutables(ctx, config, allSectionedTemplates)
	if err != nil {
		return ctx, nil, "", err
	}

	if len(substituteExecutablesFatals) > 0 {
		return ctx, nil, strings.Join(substituteExecutablesFatals, "\n\n"), nil
	}

	allSectionRows, allSectionRowsFatals := getAllSectionRows(allSectionedTemplates)
	if len(allSectionRowsFatals) > 0 {
		return ctx, nil, strings.Join(allSectionRowsFatals, "\n\n"), nil
	}

	backendInput, backendInputFatals := getBackendInput(allSectionRows, config)
	if len(backendInputFatals) > 0 {
		// Since we no longer have a sectionedTemplate errors
		// are no longer linked to a file and we separate
		// with one newline
		return ctx, nil, strings.Join(backendInputFatals, "\n"), nil
	}

	return ctx, backendInput, "", nil
}
