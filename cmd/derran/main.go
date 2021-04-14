package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"dynatrace-error-analyser/pkg/util"
	"dynatrace-error-analyser/pkg/version"
)

func main() {
	statusCode := Run(os.Args)
	os.Exit(statusCode)
}

func Run(args []string) int {
	return RunImpl(args)
}

func containsVersionFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--version" || arg == "-version" {
			return true
		}
	}

	return false
}

func RunImpl(args []string) (statusCode int) {
	if containsVersionFlag(args) {
		fmt.Println(version.ErrorAnalyser)
		return 0
	}

	statusCode = 0

	verbose, environments, specifcEnvironment, path, errorList, flagError := parseInputCommand(args)

	fmt.Println("verbose:", verbose)
	fmt.Println("environments:", environments)
	fmt.Println("specificEnvironment:", specifcEnvironment)
	fmt.Println("path:", path)
	fmt.Println("errorList:", errorList)
	fmt.Println("flagError:", flagError)

	return statusCode
}

func parseInputCommand(args []string) (verbose bool, environments map[string]string, // environment.Environment,
	specifcEnvironment string, path string, errorList []error, flagError error) {

	// define flags
	var configFile string
	var specificEnvironment string
	var versionFlag bool

	// parse flags
	shorthand := " (shorthand)"

	flagSet := flag.NewFlagSet("arguments", flag.ExitOnError)

	verboseUsage := "Set verbose flag to enable debug logging."
	flagSet.BoolVar(&verbose, "verbose", false, verboseUsage)
	flagSet.BoolVar(&verbose, "v", false, verboseUsage+shorthand)

	specificEnvironmentUsage := "Specific environment (from list) to analyse errors for."
	flagSet.StringVar(&specificEnvironment, "specific-environment", "", specificEnvironmentUsage)
	flagSet.StringVar(&specificEnvironment, "se", "", specificEnvironmentUsage+shorthand)

	configFileUsage := "Mandatory yaml file containing the environments to analyse."
	flagSet.StringVar(&configFile, "config", "", configFileUsage)
	flagSet.StringVar(&configFile, "c", "", configFileUsage+shorthand)

	versionUsage := "Prints the current version of the tool and exits"
	flagSet.BoolVar(&versionFlag, "version", false, versionUsage)

	err := flagSet.Parse(args[1:])
	if err != nil {
		return verbose, environments, specifcEnvironment, path, nil, err
	}

	// Show usage if flags are invalid
	if configFile == "" {
		cliname := "Name: \n" +
			"\tderran \n" +
			"Version:\n" +
			"\t" + version.ErrorAnalyser

		cliuse :=
			"\nUsage: \n" +
				"\tderran --config <path-to-config-yaml-file> [output-folder] \n" +
				"\tderran --config <path-to-config-yaml-file> --specific-environment <environment-name> [output-folder] \n" +
				"\tderran --version \n"

		examples :=
			"Examples:\n" +
				"Analyse Errors in all environments. Output to current folder: \n" +
				"\tderran -c='config.yaml' . \n" +
				"\nAnalyse Errors in a specific environment. Output to Temp folder: \n" +
				"\tderran --config config.yaml --specific-environment dev C:\\Temp"

		println(cliname)
		println(cliuse)
		flagSet.Usage()
		println("")
		println(examples)
		os.Exit(1)
	}

	// TODO: Delete me
	fakeEnvs := make(map[string]string)
	fakeEnvs["n1"] = "v1"
	environments, errorList = fakeEnvs, nil // environment.LoadEnvironmentList(specificEnvironment, configFile)

	path = readPath(args)

	return verbose, environments, specifcEnvironment, path, errorList, nil
}

func readPath(args []string) string {
	// Check path at the end
	potentialPath := args[len(args)-1]
	if !strings.HasSuffix(potentialPath, ".yaml") {
		potentialPath = util.ReplacePathSeparators(potentialPath)
		if _, err := ioutil.ReadDir(potentialPath); err == nil {
			if !strings.HasSuffix(potentialPath, string(os.PathSeparator)) {
				potentialPath += string(os.PathSeparator)
			}
			return potentialPath
		}
	}
	return ""
}
