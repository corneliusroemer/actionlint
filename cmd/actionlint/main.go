package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/rhysd/actionlint"
)

// do not use `const` for this variable because this variable is replaced at link time on releasing
var version = "DEV"

const usageHeader = `Usage: actionlint [FLAGS] [FILES...] [-]

  actionlint is a linter for GitHub Actions workflow files.

  To check all YAML files in current repository, just run actionlint without
  arguments. It automatically finds the nearest '.github/workflows' directory:

    $ actionlint

  To check specific files, pass the file paths as arguments:

    $ actionlint file1.yaml file2.yaml

  To check content which is not saved in file yet (e.g. output from some
  command), pass - argument. It reads stdin and checks it as workflow file:

    $ actionlint -

Configuration:

  Configuration file can be put at:

    .github/actionlint.yaml
    .github/actionlint.yml

  Please generate default configuration file and check comments in the file for
  more details.

    $ actionlint -init-config

Flags:`

func usage() {
	fmt.Fprintln(os.Stderr, usageHeader)
	flag.PrintDefaults()
}

func run(args []string, opts *actionlint.LinterOptions, initConfig bool) ([]*actionlint.Error, error) {
	l, err := actionlint.NewLinter(os.Stdout, opts)
	if err != nil {
		return nil, err
	}

	if initConfig {
		return nil, l.GenerateDefaultConfig(".")
	}

	if len(args) == 0 {
		return l.LintRepository(".")
	}

	if len(args) == 1 && args[0] == "-" {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("could not read stdin: %w", err)
		}
		return l.Lint("<stdin>", b, nil)
	}

	return l.LintFiles(args)
}

func main() {
	var ver bool
	var opts actionlint.LinterOptions
	var ignorePat string
	var initConfig bool

	flag.StringVar(&ignorePat, "ignore", "", "Regular expression matching to error messages which you want to ignore")
	flag.StringVar(&opts.Shellcheck, "shellcheck", "shellcheck", "Command name or file path of \"shellcheck\" external command")
	flag.BoolVar(&opts.Oneline, "oneline", false, "Use one line per one error. Useful for reading error messages from programs")
	flag.StringVar(&opts.ConfigFile, "config-file", "", "File path to config file")
	flag.BoolVar(&initConfig, "init-config", false, "Generate default config file at .github/actionlint.yaml in current project")
	flag.BoolVar(&opts.NoColor, "no-color", false, "Disable colorful output")
	flag.BoolVar(&opts.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&opts.Debug, "debug", false, "Enable debug output (for development)")
	flag.BoolVar(&ver, "version", false, "Show version")
	flag.Usage = usage
	flag.Parse()

	if ver {
		fmt.Println(version)
		return
	}

	if ignorePat != "" {
		r, err := regexp.Compile(ignorePat)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid regular expression %q: %s", ignorePat, err.Error())
			os.Exit(1)
		}
		opts.IgnorePattern = r
	}

	errs, err := run(flag.Args(), &opts, initConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if len(errs) > 0 {
		os.Exit(1) // Linter found some issues, yay!
	}
}
