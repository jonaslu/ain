package call

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/url"
	"os"
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

func (data Data) getBodyAsTempFile() (*os.File, error) {
	bodyStr := strings.Join(data.Body, "\n")

	// TODO Make this configurable so it can be inspected
	tmpFile, err := ioutil.TempFile("", "ain-body")
	if err != nil {
		return nil, errors.Wrap(err, "Could not create tempfile")
	}

	if _, err := tmpFile.Write([]byte(bodyStr)); err != nil {
		// This also returns an error, but the first is more significant
		// so ignore this, it's only a temp-file that will be deleted eventually
		_ = tmpFile.Close()

		return nil, errors.Wrap(err, "Could not write to tempfile")
	}

	return tmpFile, nil
}

func (data Data) runAsCurl(ctx context.Context) (string, error) {
	// TODO Put this in the global config
	args := []string{"-sS", "-vvv"}

	if data.Method != "" {
		args = append(args, "-X", strings.ToUpper(data.Method))
	}

	for _, header := range data.Headers {
		args = append(args, "-H", header)
	}

	if len(data.Body) > 0 {
		tmpFile, err := data.getBodyAsTempFile()
		if err != nil {
			return "", err
		}

		defer os.Remove(tmpFile.Name())

		args = append(args, "-d", "@"+tmpFile.Name())
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
	// TOOO Put this in the global config
	args := []string{"--ignore-stdin"}

	if data.Method != "" {
		args = append(args, strings.ToUpper(data.Method))
	}

	args = append(args, data.Host.String())

	for _, header := range data.Headers {
		args = append(args, header)
	}

	if len(data.Body) > 0 {
		tmpFile, err := data.getBodyAsTempFile()
		if err != nil {
			return "", err
		}

		defer os.Remove(tmpFile.Name())

		args = append(args, "@"+tmpFile.Name())
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
