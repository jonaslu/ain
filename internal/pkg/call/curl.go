package call

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	"github.com/pkg/errors"
)

// !! TODO !! Make this a global config
const backendTimeoutSeconds = 10

func Curl(ctx context.Context, callData *Data) (string, error) {
	args := []string{}

	for _, header := range callData.Headers {
		args = append(args, "-H")
		args = append(args, header)
	}

	args = append(args, callData.Host.String())

	curlTimeoutContext, _ := context.WithTimeout(ctx, backendTimeoutSeconds*time.Second)

	var stdout, stderr bytes.Buffer
	curlCmd := exec.CommandContext(curlTimeoutContext, "curl", args...)
	curlCmd.Stdout = &stdout
	curlCmd.Stderr = &stderr

	err := curlCmd.Run()

	stdoutStr := string(stdout.Bytes())
	stderrStr := string(stderr.Bytes())

	if err != nil {
		return "", errors.Errorf("Error: %v, running curl command: %s.\nCurl output: %s %s", curlCmd.String(), stderrStr, stdoutStr)
	}

	return stdoutStr, nil
}
