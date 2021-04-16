# Project DERRAN (Dynatrace ERRor ANalyser)
A command-line automation of the original [Dynatrace Error Analyser](https://github.com/yaypm/error-analyser).

## Goals
* To allow easy and automated analysis of Dynatrace-detected application errors
* To work across multiple SaaS and Managed environments
* To provide readable output in the form of visual reports
* To allow usage over Mission Control connections

## Intended Usage
To function properly, `derran` expects from you two YAML files:
1. Environments file (passed in with -e flag)
    * Contains the details of how `derran` should access your Dynatrace environments
2. Configurations file (passed in with -c flag)
    * Contains the details of how `derran` should analyse errors in the specified environments

### Examples:
* Analyse all errors in all environments and create a report in the current folder:

    `derran analyse --config config.yaml`

* Analyse all errors in a specific environment and create a report in Temp:

    `derran analyse -c='config.yaml' -se='dev' C:\Temp`
