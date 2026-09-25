// Command haul downloads video and audio from YouTube, X, bilibili, Xiaoyuzhou and Apple Podcasts, for people and
// for AI agents: --json prints one JSON document on stdout, it never prompts without a terminal, and exit codes
// say what went wrong.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/spf13/pflag"
)

// version is set at build time with -ldflags "-X main.version=…".
var version = "dev"

// exitUsage is for a bad command line.
const exitUsage = 64

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		console.RestoreTerminal()
	}()
	os.Exit(run(ctx, os.Args[1:]))
}

// run is haul with its arguments; it returns the exit code.
func run(ctx context.Context, args []string) int {
	cmd, rest := "download", args
	if len(args) > 0 {
		switch args[0] {
		case "download", "info", "login", "templates", "help":
			cmd, rest = args[0], args[1:]
		case "-h", "--help", "--help-hidden":
			fmt.Print(rootHelp())
			return 0
		case "--version", "-V", "version":
			fmt.Println("haul " + version)
			return 0
		}
	}
	if len(args) == 0 {
		fmt.Print(rootHelp())
		return exitUsage
	}
	switch cmd {
	case "info":
		return download(ctx, "info", rest)
	case "login":
		return login(ctx, rest)
	case "templates":
		fmt.Print(templatesHelp())
		return 0
	case "help":
		return help(rest)
	}
	return download(ctx, "download", rest)
}

func help(args []string) int {
	if len(args) == 0 {
		fmt.Print(rootHelp())
		return 0
	}
	switch args[0] {
	case "download", "info", "login", "templates":
		fmt.Print(commandHelp(args[0], false))
		return 0
	}
	return usageError(args[0], fmt.Errorf("unknown command %q", args[0]))
}

// parse reads a command's flags. It returns false with an exit code when parsing ended the run (help, or an error).
func parse(fs *pflag.FlagSet, command string, args []string) (int, bool) {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			fmt.Print(commandHelp(command, false))
			return 0, false
		}
		return usageError(command, err), false
	}
	if h, _ := fs.GetBool("help-hidden"); h {
		fmt.Print(commandHelp(command, true))
		return 0, false
	}
	if h, _ := fs.GetBool("help"); h {
		fmt.Print(commandHelp(command, false))
		return 0, false
	}
	return 0, true
}

// usageError reports a bad command line with a pointer to help, and exits 64.
func usageError(command string, err error) int {
	msg := err.Error()
	if msg != "" {
		msg = strings.ToUpper(msg[:1]) + msg[1:]
	}
	fmt.Fprintf(os.Stderr, "Error: %s\nUsage: %s\n  See 'haul %s --help' for more information.\n", msg, usageLine(command), command)
	return exitUsage
}

// fail prints an error for people and returns its exit code.
func fail(err error) int {
	console.Error(err.Error())
	return errs.KindOf(err).ExitCode()
}
