# Web Services Tool

The Web Services Tool, abbreviated as WST, is a comprehensive utility developed to facilitate the deployment and
testing of web services, although its initial focus lay with serving PHP-FPM testing needs.

While WST is tailored to be a generic tool, adaptable to a range of service testing and deployment scenarios, its
design strategy ensures it can be extended or customized for other services as needed.

## Installation

WST, a Go-based tool, can be smoothly installed from the source using the following steps:

```shell
git clone https://github.com/bukka/wst.git
cd wst
go install
```

## Usage

WST is driven by configuration and can be executed using the CLI (Command Line Interface) command, as detailed in
the subsequent sections.

### CLI

The tool provides a rich command line interface offering various commands. This can be executed as

```shell
wst [command] [options]
```

The commands supported are:

- `help` - Provides a comprehensive help list showcasing all available commands and options along with relevant
descriptions.
- `run` - Executes the predefined configuration.

Additional details for main commands are provided in the following subsections.

#### Run command

The `run` command serves as the primary engine that launches the configuration execution. Its operational steps are as
follows:

1. Construct the final configuration as outlined in the Configuration section.
2. Execute configuration actions in the specified order.
3. Execute clean up routines and terminate the operation.

As can be seen configuration and its executor are really core of the runner.

The command takes following options:

- `-c` or `--config`: This option defines the path to the configuration file. The default value points to `wst.yaml`
situated in the current working directory. If a directory is specified without a particular file, WST will default to
using the `wst.yaml` file within the provided directory. This option enables the processing of multiple configuration
files, granting higher priority to the ones defined later in the order of declaration.
- `-a` or `--all` - WST uses this option to include additional configuration files in the processing routine, even if
the `--config` option has already been specified. Particularly, it processes `wst.yaml` found in the current working
directory, `~/.wst/wst.yaml`, and `~/.config/wst/wst.yaml` if they exist.
- `-p` or `--parameter` - This option allows you to define specific parameters that can overwrite the configuration
values. It provides a way to dynamically adjust the configuration directly from the command line.
- `--no-envs` - WST, by default, checks the environment variables for any parameters that might need to be overwritten
in the configuration. Activating this option prevents environment variables from superseding the parameters defined in
the configuration files. It ensures the integrity of the configuration in environments with potentially conflicting
variable settings.
- `--dry-run` - This option activates the dry-run mode. In this mode, WST processes the configuration and performs all
preliminary setup, but refrains from executing any defined actions. This is particularly useful to verify the setup and
the operational flow without actually triggering the actions, aiding in debugging and configuration refinement.
- `--debug` - When this option is enabled, WST provides a more detailed output by logging additional debugging
information. This extensive output aids in troubleshooting by revealing the internal processing steps and any potential
issues that arise during the operation. As such, it's especially useful in development or when diagnosing problems with
WST setup.

### Configuration

The configuration, written in JSON or YAML format, encompasses all service-specific components as well as the
predefined actions for execution. The configuration structure, along with descriptions for each section, is available
in its [JsonSchema specification](schema/wst-schema.yaml). This schema is written in YAML for superior readability
and ease of editing.

