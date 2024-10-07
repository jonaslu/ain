package ain

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func NewCmdParams() *CmdParams {
	var leaveTmpFile, printCommand, showVersion, generateEmptyTemplate bool
	var envFile string

	flag.Usage = func() {
		w := flag.CommandLine.Output()

		introMsg := `Ain is an HTTP API client. It reads template files to make the HTTP call.
These can be given on the command line or sent over a pipe.

Project home page: https://github.com/jonaslu/ain`

		fmt.Fprintf(w, "%s\n\nusage: %s [options]... <template.ain>...\n", introMsg, os.Args[0])
		flag.PrintDefaults()
	}

	flag.BoolVar(&leaveTmpFile, "l", false, "Leave any temp-files")
	flag.BoolVar(&printCommand, "p", false, "Print command to the terminal instead of executing")
	flag.StringVar(&envFile, "e", ".env", "Path to .env file")
	flag.BoolVar(&showVersion, "v", false, "Show version and exit")
	flag.BoolVar(&generateEmptyTemplate, "b", false, "Generate basic template files(s)")
	flag.Parse()

	return &CmdParams{
		LeaveTmpFile:          leaveTmpFile,
		PrintCommand:          printCommand,
		ShowVersion:           showVersion,
		GenerateEmptyTemplate: generateEmptyTemplate,
		EnvFile:               envFile,
	}
}

func (c *CmdParams) SetEnvVarsAndFilenames() string {
	collectVars := false
	vars := []string{}

	for _, arg := range flag.Args() {
		if arg == "--vars" {
			collectVars = true
			continue
		}

		if collectVars {
			vars = append(vars, arg)
		} else {
			c.TemplateFileNames = append(c.TemplateFileNames, arg)
		}
	}

	for _, v := range vars {
		varName, value, found := strings.Cut(v, "=")
		if !found {
			return "Invalid environment variable format, (missing =<value>): " + v
		}
		c.EnvVars = append(c.EnvVars, []string{varName, value})
	}

	if collectVars && len(c.EnvVars) == 0 {
		return "--vars passed but no environment variables arguments found"
	}

	return ""
}

type CmdParams struct {
	LeaveTmpFile          bool
	PrintCommand          bool
	ShowVersion           bool
	GenerateEmptyTemplate bool
	EnvFile               string
	EnvVars               [][]string
	TemplateFileNames     []string
}
