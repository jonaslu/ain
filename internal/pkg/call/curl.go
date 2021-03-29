package call

import (
	"context"
	"os/exec"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
)

type curl struct {
	args []string
}

func newCurlBackend(data *data.Parse) (*curl, error) {
	args := data.BackendOptions

	if data.Method != "" {
		args = append(args, "-X", strings.ToUpper(data.Method))
	}

	for _, header := range data.Headers {
		args = append(args, "-H", header)
	}

	if len(data.Body) > 0 {
		tmpFile, err := data.GetBodyAsTempFile()
		if err != nil {
			return nil, err
		}

		// defer os.Remove(tmpFile.Name())

		args = append(args, "-d", "@"+tmpFile.Name())
	}

	args = append(args, data.Host.String())

	return &curl{args: args}, nil
}

func (curl curl) runAsCmd(ctx context.Context) ([]byte, error) {
	curlCmd := exec.CommandContext(ctx, "curl", curl.args...)
	return curlCmd.CombinedOutput()
}
