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

The WST (Web Services Tool) command-line interface operates based on the following structure:

```shell
wst [global_options] [command] [command_options]
```

In the absence of any provided commands or options, a help list is displayed. This list outlines available commands and
options coupled with relevant descriptions.

Several global options are present:

- `-h` or `--help` - Triggers the display of the help list.
- `--version` - This option prints the version number of WST. It follows semantic versioning conventions.
- `--debug` - When this option is set, WST provides a more detailed output by logging additional debugging information. 
This extensive output aids in troubleshooting by revealing the internal processing steps and any potential issues that 
arise during the operation. As such, it is especially useful in development or when diagnosing problems with WST setup.
All subcommands inherit this option, therefore, it remains activated during the entire course of the command execution.

The commands supported are:

- `help` - Triggers the display of the help list.
- `run` - Executes the predefined configuration.

Additional details for the run command are provided in the following subsection.

#### Run command

The `run` command serves as the primary engine that launches the configuration execution. Its operational steps are as
follows:

1. Construct the final configuration as outlined in the Configuration section.
2. Execute configuration actions in the specified order.
3. Execute clean up routines and terminate the operation.

As can be seen configuration and its executor are really core of the runner.

The command takes following options:

- `-c` or `--config`: This option defines the path to the configuration file. The default value points to `wst.yaml`
in the current working directory. If a directory is specified without a file, WST will default to
using the `wst.yaml` file within the provided directory. It is possible to specify this option multiple times to process
multiple configuration files, granting higher priority to the ones defined later in the order of declaration.
- `-a` or `--all` - WST uses this option to include additional configuration files in the processing routine, even if
the `--config` option has already been specified. Particularly, it processes `wst.yaml` found in the current working
directory, `~/.wst/wst.yaml`, and `~/.config/wst/wst.yaml` if they exist.
- `-o` or `--overwrite` - This option allows you to define specific values that can overwrite the configuration
values. It provides a way to dynamically adjust the configuration directly from the command line. The overwrite value
is composed of `key=value` string where `key` is the config position in dot notation and the `value` is the actual
value to overwrite it with. For example to overwrite the name of the first instance, it could be done using
`spec.instances[0].name=new name`.
- `--no-envs` - WST, by default, checks the environment variables for WST customization. Activating this option
prevents environment variables from being checked.
- `--dry-run` - This option activates the dry-run mode. In this mode, WST processes the configuration and performs all
preliminary setup, but refrains from executing any defined actions. This is particularly useful to verify the setup and
the operational flow without actually triggering the actions, aiding in debugging and configuration refinement.

As highlighted in the options description, the application also checks the environment variables. Currently only
`WST_OVERWRITE` is supported, which enables overwriting of the configuration. It supports the same format as in
`--overwrite` option value, but it also allows providing multiple such values separated by a colon. For example, if one
wants to change the name and nginx service sandbox of the first instance, it could be done as follows:
```bash
WST_OVERWRITE='spec.instances[0].name=new name:spec.instances[0].services.nginx.sandbox=docker'
```

### Configuration

The configuration, written in JSON or YAML format, encompasses all service-specific components as well as the
predefined actions for execution. The configuration structure, along with descriptions for each section, is available
in its [JsonSchema specification](schema/wst-schema.yaml). This schema is written in YAML for superior readability
and ease of editing.

### Architecture

### Execution Architecture

The execution is designed with flexibility and scalability in mind, accommodating various execution contexts through
a structured yet adaptable architecture. This architecture is built upon four key entities: Sandboxes, Servers,
Services, and Environments. Each plays a crucial role in ensuring that our application can be deployed and managed
effectively across different platforms.

#### Sandboxes

A Sandbox represents the fundamental execution unit within our architecture. It is where individual instances
of services are executed, isolated from each other to ensure security and stability. Sandboxes can be categorized
based on their execution context:

- **Local**: Execution as a standalone process within a host operating system.
- **Docker**: Deployment as a containerized service, utilizing Docker for virtualization.
- **Kubernetes**: Deployment within a Kubernetes cluster, managed as a deployment.

#### Servers

A Server defines parametrized configurations, definitions, and specific hooks tailored to each Sandbox. It encapsulates
all necessary information and operations required to manage a service's lifecycle, such as starting, stopping, and
restarting. For instance, a Server designed for Nginx would include configuration file templates, Docker image
details, and commands to manage the service across different Sandboxes.

#### Services

A Service represents the actual execution instance of a Server, complete with all required configuration parameters and
execution information. This design allows a single Server to be instantiated as multiple Services, each with 
potentially unique configurations. Services are executed within Sandboxes, leveraging the Server's definitions
to ensure consistent management and operation.

#### Environments

An Environment is a higher-level abstraction that groups together Sandboxes of a specific type, facilitating
communication and interaction among services deployed within them. It serves as a shared context or domain where:

- For **Local** execution, it might represent a specific filesystem root where configurations are stored.
- In **Docker**, it could imply a shared network that allows containerized services to communicate.
- For **Kubernetes**, it generally encompasses a namespace, providing a unified operational scope for deployments.

The concept of an Environment is essential for managing and scaling our application's deployment across different 
execution contexts. It allows for a cohesive management layer that abstracts the underlying complexities of service 
interaction and network communication, ensuring that services can seamlessly discover and communicate with each other
regardless of their deployment context.
