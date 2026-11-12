// Command dia-config validates Dia application configuration before optionally
// preparing reviewable deployment artifacts. It never executes a deployment.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/mejango/blossm-toolbox-go/cli/blossm"
	"github.com/mejango/dia-go/deployment"
)

func run(args []string, out, stderr io.Writer) error {
	flags := flag.NewFlagSet("dia-config", flag.ContinueOnError)
	flags.SetOutput(stderr)
	path := flags.String("config", "", "Dia application configuration path (required)")
	directory := flags.String("out", "", "optional directory for prepared artifacts; omission validates only")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *path == "" || flags.NArg() != 0 {
		return errors.New("usage: dia-config --config path [--out directory]")
	}
	c, err := deployment.ReadConfig(*path)
	if err != nil {
		return err
	}
	if *directory == "" {
		_, err = fmt.Fprintln(out, "configuration valid")
		return err
	}
	plan, err := deployment.Prepare(c)
	if err != nil {
		return err
	}
	if err = cli.WritePlan(*directory, plan); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, "deployment artifacts prepared; no commands executed")
	return err
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
