package call

import (
	"bytes"
	"context"
	"net/url"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type Data struct {
	Host       *url.URL
	URL        string
	Body       []string
	Parameters []string
	Method     string
	Headers    []string
}

func (data Data) runAsCurl(ctx context.Context) (string, error) {
	args := []string{"-sS"}

	if data.Method != "" {
		args = append(args, "-X", strings.ToUpper(data.Method))
	}

	for _, header := range data.Headers {
		args = append(args, "-H", header)
	}

	args = append(args, data.Host.String())

	var stdout, stderr bytes.Buffer
	curlCmd := exec.CommandContext(ctx, "curl", args...)
	curlCmd.Stdout = &stdout
	curlCmd.Stderr = &stderr

	err := curlCmd.Run()
	stdoutStr := string(stdout.Bytes())

	if err != nil {
		stderrStr := string(stderr.Bytes())
		return "", errors.Errorf("Error: %v, running curl: %s.\nError output: %s %s", err, curlCmd.String(), stderrStr, stdoutStr)
	}

	return stdoutStr, nil

}

func (data Data) runAsHttpie(ctx context.Context) (string, error) {
	args := []string{"--ignore-stdin"}

	if data.Method != "" {
		args = append(args, strings.ToUpper(data.Method))
	}

	args = append(args, data.Host.String())

	for _, header := range data.Headers {
		args = append(args, header)
	}

	var stdout, stderr bytes.Buffer
	httpCmd := exec.CommandContext(ctx, "http", args...)
	httpCmd.Stdout = &stdout
	httpCmd.Stderr = &stderr

	err := httpCmd.Run()
	stdoutStr := string(stdout.Bytes())

	if err != nil {
		stderrStr := string(stderr.Bytes())
		return "", errors.Errorf("Error: %v, running http: %s.\nError output: %s %s", err, httpCmd.String(), stderrStr, stdoutStr)
	}

	return stdoutStr, nil
}
