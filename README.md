# Web Services Tool

WST (Web Services Tool) is a tool for web service deployment and testing.

The tool was developed for PHP-FPM testing, but there is nothing PHP-FPM specific as the tool is meant to be generic
and potentially usable for testing and deploying other services.

## Installation

The tool is written in go and can be installed from source like

```shell
git clone https://github.com/bukka/wst.git
cd wst
go install
```

## Usage

The tool is driven by configuration. The tool is executed using CLI command as documented below.

### CLI

The tool provides a rich command line interface offering various commands. This can be executed as

```shell
wst [command] [options]
```

The supported commands are following:

- help - displays detailed help listing all commands and their actions as well as their description
- run - executes the config
- verify - verifies the config

The following subsection further document some of the main commands

#### Run command

The `run` command is the primary command that executes the configuration. Specifically it does following things:

1. Create final configuration as documented in Configuration section
2. Execute configuration actions in the specified order
3. Clean up and terminate

As can be seen configuration and its executor are really core of the runner.

### Configuration

The configuration needs to be written in YAML or JSON format and specifies all service specific parts as well as
actions to execute. The configuration structure with description of all parts can be found in its
[JsonSchema specification](schema/wst-schema.yaml) which is written in YAML for easier reading and editing.

