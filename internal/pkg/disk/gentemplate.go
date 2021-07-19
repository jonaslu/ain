package disk

import (
	"flag"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

var basicTemplate = `[Host]
http://localhost:${PORT}

[Headers]
Content-Type: application/json

# [Method]
# POST

# [Body]
# {
#    "some": "json"
# }

[Config]
Timeout=3

[Backend]
curl
# httpie

[BackendOptions]
-sS

# Short help:
# Comments start with hash-sign (#) and are ignored.
# ${VARIABLES} are replaced with the .env-file or environment variable value
# $(executables.sh) are replaced with the output of that executable`

func GenerateEmptyTemplates() error {
	var templateFileNames []string

	if len(flag.Args()) >= 1 {
		templateFileNames = flag.Args()
	}

	if len(templateFileNames) == 0 {
		// Write to STDOUT
		_, err := fmt.Fprintln(os.Stdout, basicTemplate)
		return err
	}

	for _, filename := range templateFileNames {
		_, err := os.Stat(filename)

		if !os.IsNotExist(err) {
			return errors.Errorf("Cannot write basic template. File already exists %s", filename)
		}

		err = os.WriteFile(filename, []byte(basicTemplate), 0644)

		if err != nil {
			return errors.Wrapf(err, "Could not write basic template to file %s", filename)
		}
	}

	return nil
}
