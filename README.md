# Project DERRAN (Dynatrace ERRor ANalyser)

This is a command-line tool which automates error analysis in Dynatrace environments.

The purpose of the tool is to be able to analyse errors detected by Dynatrace across multiple environments applying different use cases to understand the monetary impact these errors have on your business. Reports are then produced to illustrate any findings.

#### Objectives

* To allow easy and automated analysis of Dynatrace-detected application errors
* To work across multiple SaaS and Managed environments
* To provide readable output in the form of visual reports
* To allow usage over Mission Control connections

#### You may also want to check out:
* [Dynatrace Error Analyser](https://github.com/yaypm/error-analyser)
  * This is the original project that this tool is based on. 
  * The data analysis functionality is the same, however, it offers a more guided, user friendly, GUI-based experience
* [Dynatrace Monitoring as Code](https://github.com/dynatrace-oss/dynatrace-monitoring-as-code) (monaco) 
  * This project heavily inspired the development of this tool
  * While `derran` can only perform data analysis and reporting tasks, `monaco` can help you manage Dynatrace configuration at scale, with a configuration-as-code approach

#### Table of contents
- [Project DERRAN (Dynatrace ERRor ANalyser)](#project-derran-dynatrace-error-analyser)
      - [Objectives](#objectives)
      - [You may also want to check out:](#you-may-also-want-to-check-out)
      - [Table of contents](#table-of-contents)
  - [Installation](#installation)
      - [On Mac or Linux systems perform the following](#on-mac-or-linux-systems-perform-the-following)
      - [On Windows](#on-windows)
  - [Basic Usage](#basic-usage)
      - [Examples:](#examples)
  - [File requirements](#file-requirements)
      - [Environments file](#environments-file)
      - [Config file](#config-file)
  - [Reporting](#reporting)

---

## Installation

For convenience, ```derran``` has been already packaged into a binary which can be downloaded from the [releases section](https://github.com/radu-stefan-dt/dynatrace-error-analyser/releases) of this repository. Find the binary most appropriate for your system, rename it to 'derran' and add it to your system PATH.
Adding a binary to PATH will differ depending on the Operating System.

#### On Mac or Linux systems perform the following

Print a colon-separated list of locations in your ```PATH```:

```
$ echo $PATH
```

Move the derran binary to one of the listed locations. This command assumes that the binary is currently in your downloads folder and that your PATH includes ```/usr/local/bin```:

```
mv ~/Downloads/derran /usr/local/bin/
```

#### On Windows

From the user interface, use [this Stack OverFlow guide](https://stackoverflow.com/questions/1618280/where-can-i-set-path-to-make-exe-on-windows) to set the PATH on Windows.

Verify the installation by running ```derran``` from your terminal.

---

## Basic Usage

`derran` is a command line tool. It runs as a single command with mandatory and optional arguments such as `--environments`, `--config`, or `--dry-run`. As `derran` will produce reports based on the findings, it requires an output folder as the last positional argument. If none is specified, the current working directory is used.

The following are currently supported flags for the `analyse` command:
```
--verbose, -v                             (default: false)
--dry-run, -d                             validate connectivity to environments, but don't execute anything (default: false)
--environments value, -e value            YAML file containing details of Dynatrace environments
--config value, -c value                  YAML file containing configurations for error analysis
--specific-environment value, --se value  Specific environment (from list) to analyse
--help, -h                                show help (default: false)
```
Running `derran` is done with mandatory and optional options and positional arguments:
```
derran --environments <path-to-environments-file> --config <path-to-configuration-file> [--specific-environment <environment-name>] [--dry-run] [--verbose] [report-output-folder]
```
#### Examples:
* Analyse all errors in all environments and create reports in the current folder:
    ```
    derran analyse --environments envs.yaml --config config.yaml
    ```
* Analyse all errors in a specific environment and create reports in C:\Temp:
    ```
    derran analyse -e='envs.yaml' -c='config.yaml' -se='dev' C:\Temp
    ```

---

## File requirements

To function properly, `derran` expects from you two YAML files:
1. Environments file
    * Contains the details of how `derran` should access your Dynatrace environments
2. Configurations file
    * Contains the details of how `derran` should analyse errors in the specified environments

#### Environments file

This file is mandatory for running the tool and it's passed in with flag ```--environments, -e```. It must be an absolute or relative path to a YAML file. The contents define all the environments `derran` should connect to and how to do so.

To define an environment, give it a unqiue ID as the key and list out the following mandatory attributes:

- **name**
  - represents the name by which to refer to this environment
- **env-url**
  - represents the URL at which this environment can be reached
- **env-token-name** or **env-token**
  - only one of these should be specify as it represents the token to use in authenticating the requests to the Dynatrace API
  - **env-token-name** expects the name of an Environment Variable which holds the token value. This is the preferred option here as it increases security by not relying on this information to be present in plain text in the file.
  - For the less technical users or those with limited priviledges on their system, **env-token** allows you to specify the token value directly in your YAML file.

> If your only connectivity to an environment is through Dynatrace Mission Control, you can specify two additional attributes: **mc-ua** and **mc-cookie**. An Dynatrace internal guide is provided at this link if you are unsure how to use these attributes and what data they require..

An environments file may end up looking like this:
```yaml
foo:
    - name: "foo"
    - env-url: "https://foo.live.dynatrace.com"
    - env-token-name: "FOO_TOKEN"

bar:
    - name: "bar"
    - env-url: "https://bar.dynatrace-managed.com/e/abcdef123456"
    - env-token: "abcdefghijkl123456789"

foobar:
    - name: "foobar"
    - env-url: "https://foobar.dynatrace-managed.com/e/xxxxxxxxx"
    - env-token-name: "FOOBAR_ENV_TOKEN"
    - mc-ua: "FOOBAR_MC_UA"
    - mc-cookie: "FOOBAR_MC_COOKIE"
```

#### Config file

This file is mandatory for running the tool and is passed in with flag ```--config, -c```. It must be an absolute or relative path to a YAML file. The contents define all configurations `derran` should use when analysing errors in your Dynatrace environments.

To define a configuration, give it a unique ID and list out the following mandatory attributes:
- **name**
  - Represents a name or description for this configuration. Use it to distinguish between multiple configurations in a human-readable way.
- **use_cases**
  - Given as a list of one or more of: `lost_basket`, `agent_hours`, or `incurred_costs`.
  - Represents the use case(s) by which `derran` should apply business impact calculations. This will dictate what properties will be madatory in the next  section.
    - Lost orders - quantify the revenue at risk from users that have not converted (and did not return) after encountering an error.
    - Agent hours - estimate the hours (and cost, where relevant) to be spent by support staff, dealing with enquiries due to an error.
    - Incurred costs - total up the potential costs incurred from an error hitting a user, for example a penalty due to an SLA.
- **properties**
  - Represents mandatory and optional properties that `derran` will use to extract and analyse Dynatrace data. Mandatory properties will differ depending on the selected use cases.
  - For all use cases:
    - **error_prop** (mandatory) - represents a Dynatrace Session Property which captures the title of an error, storede as a string
    - **conversion** (mandatory) - represents the name of a Dynatrace User Action which marks a converted session
    - **application** (optional) - represents the display name of a Dynatrace Application and is used to filter the data and results to one application. Otherwise, the configuration is applied across all RUM Applications in the Dynatrace environment.
  - For `lost_basket` use case:
    - **basket_prop** (mandatory) - reprsents a Dynatrace Session Property which captures a user's order (or basket) value, stored as a double.
    - **margin** (optional) - represents the profit margin (as a percentage) by which to calculate the true cost lost to the business. 
    - **multiplication** (optional) - represents a multiplication factor. In the case that users are purchasing a subscription-based service (e.g. phone or broadband) we will only capture the initial cost; this factor can then be set to e.g. the minimum contract length to get a more accurate view of total revenue loss.
  - For `agent_hours` use case:
    - **users_calling_in** (mandatory) - represents the percentage of users who are likely to call in and report an IT issue
    - **length_of_call** (mandatory) - represents the average length of a support call related to IT errors
    - **cost_of_call** (optional) - represents the average cost of a call into the call centre when a user needs support
  - For `incurred_costs` use case:
    - **cost_of_error** (mandatory) - represents the average costs incurred when an error occurs
- **environments**
  - Represents a list of Dynatrace Environments (defined in your environments file) that this configuration should be applied to.
  - The environments should be referenced by name

A config file may end up looking like this:
```yaml
foo:
    name: "My lost baskets config"
    use_cases:
        - "lost_basket"
    properties:
        application: "FooApp"
        error_prop: "fooerror"
        conversion: "click on foo on page bar"
        basket_prop: "foorevenue"
    environments:
        - "foo"
        - "bar"

bar:
    name: "Config for environment foobar"
    use_cases:
        - "agent_hours"
        - "incurred_costs"
    properties:
        application: "BarApp"
        error_prop: "errorbox"
        conversion: "loading of page foobar"
        users_calling_in: 30
        length_of_call: 25
        cost_of_call: 12
        cost_of_error: 350
    environments:
        - "foobar"
```

---

## Reporting