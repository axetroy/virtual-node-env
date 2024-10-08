package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/axetroy/nodapt/internal/cli"
	"github.com/axetroy/nodapt/internal/node"
	"github.com/axetroy/nodapt/internal/util"
	"github.com/pkg/errors"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func printHelp() {
	fmt.Println(`nodapt - A virtual node environment for node.js, node version manager for projects.

USAGE:
  nodapt [OPTIONS] <ARGS...>
  nodapt [OPTIONS] run <ARGS...>
  nodapt [OPTIONS] use <CONSTRAINT> <ARGS...>
  nodapt [OPTIONS] rm <CONSTRAINT> <ARGS...>
  nodapt [OPTIONS] clean
  nodapt [OPTIONS] ls
  nodapt [OPTIONS] ls-remote

COMMANDS:
  run <ARGS...>               Automatically select node version to run commands
  use <CONSTRAINT> <ARGS...>  Use the specified version of node to run the command
  rm|remove <CONSTRAINT>      Remove the specified version of node that installed by nodapt
  clean                       Remove all the node version that installed by nodapt
  ls|list                     List all the installed node version
  ls-remote|list-remote       List all the available node version
  <ARGS...>                   Alias for 'run <ARGS...>' but shorter

OPTIONS:
  --help|-h                   Print help information
  --version|-v                Print version information

ENVIRONMENT VARIABLES:
  NODE_MIRROR                 The mirror of the nodejs download, defaults to: https://nodejs.org/dist/
                              Chinese users defaults to: https://registry.npmmirror.com/-/binary/node/
  NODE_ENV_DIR                The directory where the nodejs is stored, defaults to: $HOME/.nodapt
  DEBUG                       Print debug information when set DEBUG=1

EXAMPLES:
  nodapt node -v
  nodapt run node -v
  nodapt use v14.17.0 node -v

SOURCE CODE:
  https://github.com/axetroy/nodapt`)
}

type Flag struct {
	IsPrintHelp    bool
	IsPrintVersion bool
	Cmd            []string
}

func parse() (*Flag, error) {
	args := os.Args[1:]

	f := Flag{}

	length := 0
	commandIndex := -1

	for length < len(args) {
		arg := args[length]

		if commandIndex >= 0 && length > commandIndex {
			f.Cmd = append(f.Cmd, arg)
			length++
			continue
		}

		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			switch true {
			case arg == "--help", arg == "-h":
				f.IsPrintHelp = true
			case arg == "--version", arg == "-v":
				f.IsPrintVersion = true
			case arg == "--":
				if commandIndex == -1 {
					commandIndex = length
				} else {
					f.Cmd = append(f.Cmd, arg)
				}
			}
		} else {
			commandIndex = length
			f.Cmd = append(f.Cmd, arg)
		}

		length++
	}

	return &f, nil
}

func run() error {
	cwd, err := os.Getwd()

	if err != nil {
		return errors.WithStack(err)
	}

	f, err := parse()
	if err != nil {
		return errors.WithStack(err)
	}

	if f.IsPrintHelp {
		printHelp()
		os.Exit(0)
	}

	if f.IsPrintVersion {
		fmt.Printf("%s %s %s\n", version, commit, date)
		os.Exit(0)
	}

	if len(f.Cmd) == 0 {
		return errors.New("missing command")
	}

	cmd := f.Cmd[0]

	switch cmd {
	case "use":
		if len(f.Cmd) == 1 {
			return fmt.Errorf("constraint is required")
		}

		constraint := f.Cmd[1]
		commands := f.Cmd[2:]

		if len(commands) == 0 {
			return cli.Use(constraint)
		} else {

			return cli.RunWithConstraint(constraint, commands)
		}
	case "rm", "remove":
		if len(f.Cmd) == 1 {
			return fmt.Errorf("constraint is required")
		}

		constraint := f.Cmd[1]

		return cli.Remove(constraint)

	case "ls", "list":
		return cli.List()
	case "ls-remote", "list-remote":
		return cli.ListRemote()
	case "clean":
		return cli.Clean()
	case "run":
		f.Cmd = f.Cmd[1:]
		fallthrough
	default:
		if len(f.Cmd) == 0 {
			return errors.New("commands is required")
		}

		packageJSONPath := util.LoopUpFile(cwd, "package.json")

		// If the package.json file is found, then use the node constraint in the package.json to run the command
		if packageJSONPath != nil {
			util.Debug("Use node constraint from %s\n", *packageJSONPath)

			constraint, err := node.GetConstraintFromPackageJSON(*packageJSONPath)

			if err != nil {
				return errors.WithMessagef(err, "failed to get node constraint from %s", *packageJSONPath)
			}

			if constraint != nil {
				util.Debug("Use node constraint: %s\n", *constraint)
				return cli.RunWithConstraint(*constraint, f.Cmd)
			} else {
				util.Debug("Run command directly\n")
				return cli.RunDirectly(f.Cmd)
			}
		} else {
			util.Debug("Run command directly\n")
			return cli.RunDirectly(f.Cmd)
		}
	}
}

func main() {
	if err := run(); err != nil {
		if os.Getenv("DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
			fmt.Fprintf(os.Stderr, "current commit hash %s\n", commit)
		} else {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			fmt.Fprintln(os.Stderr, "Print debug information when set DEBUG=1")
		}

		unwrapError := errors.Unwrap(err)

		if err, ok := err.(*exec.ExitError); ok {
			os.Exit(err.ExitCode())
		} else if err, ok := unwrapError.(*exec.ExitError); ok {
			os.Exit(err.ExitCode())
		} else {
			os.Exit(1)
		}
	}
}
