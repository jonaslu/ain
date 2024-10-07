package disk

import (
	"io"
	"os"

	"github.com/jonaslu/ain/internal/pkg/utils"
	"github.com/pkg/errors"
)

func GetTemplateFilenames(cmdParamTemplateFileNames []string) ([]string, error) {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "Could not stat stdin")
	}

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		fileNameBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, errors.Wrap(err, "Could not read pipe stdin")
		}

		localTemplateFilenamesViaPipe, err := utils.TokenizeLine(string(fileNameBytes))
		if err != nil {
			return nil, errors.Wrap(err, "Could not read template name(s) from pipe")
		}

		if len(localTemplateFilenamesViaPipe) == 0 {
			return nil, errors.New("Pipe input did not contain any template names")
		}

		cmdParamTemplateFileNames = append(cmdParamTemplateFileNames, localTemplateFilenamesViaPipe...)
	}

	return cmdParamTemplateFileNames, nil
}
