package disk

import (
	"flag"
	"io/ioutil"
	"os"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/utils"
	"github.com/pkg/errors"
)

func GetTemplateFilenames() ([]string, error) {
	var localTemplateFilenames []string

	if len(flag.Args()) >= 1 {
		localTemplateFilenames = flag.Args()
	}

	fi, err := os.Stdin.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "could not stat stdin")
	}

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		// Connected to a pipe
		fileNameBytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, errors.Wrap(err, "could not read stdin")
		}

		localTemplateFilenamesViaPipe, err := utils.TokenizeLine(string(fileNameBytes), true)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse filenames from pipe")
		}

		localTemplateFilenames = append(localTemplateFilenames, localTemplateFilenamesViaPipe...)
	}

	trimmedLocalTemplateFilenames := []string{}
	for _, localTemplateFilename := range localTemplateFilenames {
		trimmedLocalTemplateFilenames = append(trimmedLocalTemplateFilenames, strings.TrimSpace(localTemplateFilename))
	}

	return trimmedLocalTemplateFilenames, nil
}
