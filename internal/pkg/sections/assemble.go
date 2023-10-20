package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/disk"
)

func Assemble(ctx context.Context, filenames []string) (*data.BackendInput, string, error) {
	allSectionedTemplates := []*SectionedTemplate{}
	fatals := []string{}

	for _, filename := range filenames {
		// !! TODO !! The file-name will be displayed as test.ain! <- Remove the exclamation-mark
		// when setting the file-name2.
		rawTemplateString, err := disk.ReadRawTemplateString(filename)
		if err != nil {
			return nil, "", err
		}

		if sectionedTemplate := NewSections(rawTemplateString, filename); sectionedTemplate.HasFatalMessages() {
			fatals = append(fatals, sectionedTemplate.GetFatalMessages())
		} else {
			allSectionedTemplates = append(allSectionedTemplates, sectionedTemplate)
		}
	}

	if len(fatals) > 0 {
		return nil, strings.Join(fatals, "\n\n"), nil
	}

	for _, sectionedTemplate := range allSectionedTemplates {
		if sectionedTemplate.substituteEnvVars(); sectionedTemplate.HasFatalMessages() {
			fatals = append(fatals, sectionedTemplate.GetFatalMessages())
		}
	}

	if len(fatals) > 0 {
		return nil, strings.Join(fatals, "\n\n"), nil
	}

	config := data.NewConfig()

	for i := len(allSectionedTemplates) - 1; i >= 0; i-- {
		sectionedTemplate := allSectionedTemplates[i]
		localConfig := sectionedTemplate.getConfig()

		if sectionedTemplate.HasFatalMessages() {
			fatals = append(fatals, sectionedTemplate.GetFatalMessages())
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

	if len(fatals) > 0 {
		return nil, strings.Join(fatals, "\n\n"), nil
	}

	allExecutableAndArgs := []executableAndArgs{}
	for _, sectionedTemplate := range allSectionedTemplates {
		allExecutableAndArgs = append(allExecutableAndArgs, sectionedTemplate.captureExecutableAndArgs()...)

		if sectionedTemplate.HasFatalMessages() {
			fatals = append(fatals, sectionedTemplate.GetFatalMessages())
		}
	}

	if len(fatals) > 0 {
		return nil, strings.Join(fatals, "\n\n"), nil
	}

	allExecutablesOutput := callExecutables(ctx, config, allExecutableAndArgs)

	for _, sectionedTemplate := range allSectionedTemplates {
		if sectionedTemplate.insertExecutableOutput(&allExecutablesOutput); sectionedTemplate.HasFatalMessages() {
			fatals = append(fatals, sectionedTemplate.GetFatalMessages())
		}
	}

	if len(fatals) > 0 {
		return nil, strings.Join(fatals, "\n\n"), nil
	}

	var host, backend, method string
	var headers, query, body []string
	var backendOptions [][]string

	for _, sectionedTemplate := range allSectionedTemplates {
		host = host + sectionedTemplate.getHost()
		headers = append(headers, sectionedTemplate.getHeaders()...)
		query = append(query, sectionedTemplate.getQuery()...)
		backendOptions = append(backendOptions, sectionedTemplate.getBackendOptions()...)

		if localBackend := sectionedTemplate.getBackend(); localBackend != "" {
			backend = localBackend
		}

		if localMethod := sectionedTemplate.getMethod(); localMethod != "" {
			method = localMethod
		}

		if localBody := sectionedTemplate.getBody(); len(localBody) > 0 {
			body = localBody
		}

		if sectionedTemplate.HasFatalMessages() {
			fatals = append(fatals, sectionedTemplate.GetFatalMessages())
		}
	}

	if len(fatals) > 0 {
		return nil, strings.Join(fatals, "\n\n"), nil
	}

	var backendInput data.BackendInput

	if host == "" {
		fatals = append(fatals, "No mandatory [Host] section found")
	} else {
		hostUrl, err := url.Parse(host)

		if err != nil {
			fatals = append(fatals, fmt.Sprintf("[Host] has illegal url: %s, error: %v", host, err))
		} else {
			addQueryString(hostUrl, query, config)
			backendInput.Host = hostUrl
		}
	}

	if backend == "" {
		fatals = append(fatals, "No mandatory [Backend] section found")
	}

	if len(fatals) > 0 {
		// Since we no longer have a sectionedTemplate errors
		// are no longer linked to a file and we separate
		// with one newline
		return nil, strings.Join(fatals, "\n"), nil
	}

	spew.Dump(backendInput, method, body)

	return &backendInput, "", nil
}

func main() {
	flag.Parse()
	filenames := flag.Args()

	_, fatals, err := Assemble(context.TODO(), filenames)
	if err != nil {
		panic(err)
	}

	if fatals != "" {
		fmt.Println(fatals)
	}
}
