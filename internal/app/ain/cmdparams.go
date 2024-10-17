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

		fmt.Fprintf(w, "%s\n\nusage: %s [OPTIONS] <template.ain> [--vars VAR=VALUE ...] \n", introMsg, os.Args[0])
		fmt.Fprintf(w, "\nOPTIONS:\n")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "  -%-22s %s\n", f.Name, f.Usage)
		})

		fmt.Fprintf(w, "\nARGUMENTS:\n")
		fmt.Fprintf(w, "  <template.ain>[!]       One or more template files to process. Required\n")
		fmt.Fprintf(w, "  --vars VAR=VALUE [...]  Values for environment variables, set after <template.ain> file(s)\n")
	}

	flag.BoolVar(&leaveTmpFile, "l", false, "Leave any body-files")
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

func (c *CmdParams) SetEnvVarsAndFilenames() error {
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
			return fmt.Errorf("invalid environment variable format, (missing =<value>): %s", v)
		}
		c.EnvVars = append(c.EnvVars, []string{varName, value})
	}

	if collectVars && len(c.EnvVars) == 0 {
		return fmt.Errorf("--vars passed but no environment variables arguments found")
	}

	return nil
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
