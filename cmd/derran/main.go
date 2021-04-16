package main

import (
	"fmt"
	"os"

	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/analyse"
	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/util"
	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/version"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
)

func main() {
	statusCode := Run(os.Args)
	os.Exit(statusCode)
}

func Run(args []string) int {
	return RunImpl(args, afero.NewOsFs())
}

func RunImpl(args []string, fs afero.Fs) (statusCode int) {
	var app *cli.App = buildCli(fs)

	if err := app.Run(args); err != nil {
		util.Log.Error("%s\n", err)
		return 1
	}

	return 0
}

func buildCli(fs afero.Fs) *cli.App {
	app := cli.NewApp()

	app.Usage = "Automates the impact analysis of application errors detected by Dynatrace."
	app.Version = version.ErrorAnalyser

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(c.App.Version)
	}

	cli.VersionFlag = &cli.BoolFlag{
		Name:  "version",
		Usage: "print the version and exit",
	}

	app.Description = `
	Tool used to analyse and report on application errors via the CLI

	Examples:
	  Analyse all errors in all environments and create a report in the current folder:
	    derran analyse --config config.yaml .

	  Analyse all errors in a specific environment and create a report in Temp:
	    derran analyse -c='config.yaml' -se='dev' C:\Temp
	`
	analyseCommand := getAnalyseCommand(fs)
	app.Commands = []*cli.Command{&analyseCommand}

	return app
}

func getAnalyseCommand(fs afero.Fs) cli.Command {
	command := cli.Command{
		Name:      "analyse",
		Usage:     "analyses errors in given environments",
		UsageText: "analyse [command options] [output directory]",
		ArgsUsage: "[output directory]",
		Before: func(c *cli.Context) error {
			if err := util.SetupLogging(c.Bool("verbose")); err != nil {
				return err
			}

			util.Log.Info("Dynatrace Error Analyser v" + version.ErrorAnalyser)

			return nil
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
			},
			&cli.PathFlag{
				Name:      "config",
				Usage:     "Yaml file contianing the tool configuration",
				Aliases:   []string{"c"},
				Required:  true,
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:    "specific-environment",
				Usage:   "Specific environment (from list) to analyse",
				Aliases: []string{"se"},
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() > 1 {
				fmt.Println("Error: Too many arguments! Either specify a relative path to the working directory, or omit it for using the current working directory.")
				cli.ShowAppHelpAndExit(ctx, 1)
			}

			var outputDir string

			if ctx.Args().Present() {
				outputDir = ctx.Args().First()
			} else {
				outputDir = "."
			}

			return analyse.Analyse(
				outputDir,
				fs,
				ctx.Path("config"),
				ctx.String("specific-environment"),
			)
		},
	}
	return command
}
