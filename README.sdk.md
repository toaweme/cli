# SDK Reference

## Core interfaces

The app implements:

{{ file "docs/templates/inputs/sdk_interfaces_app_interface.md" }}

Every command implements:

{{ file "docs/templates/inputs/sdk_interfaces_command_interface.md" }}

Commands can optionally implement:

{{ file "docs/templates/inputs/sdk_interfaces_example_provider_interface.md" }}

## Types

{{ file "docs/templates/inputs/sdk_types_settings_type.md" }}

{{ file "docs/templates/inputs/sdk_types_global_options_type.md" }}

{{ file "docs/templates/inputs/sdk_types_unknowns_type.md" }}

{{ file "docs/templates/inputs/sdk_types_base_command_type.md" }}

## Struct tags

Command config structs use these tags:

{{ file "docs/templates/inputs/sdk_struct_tags_tags_table.md" }}

Tags combine freely on a single field:

{{ file "docs/templates/inputs/sdk_struct_tags_combined_example.md" }}

## Creating an app

{{ file "docs/templates/inputs/sdk_creating_app_example.md" }}

`NewApp` returns `*CLI` which satisfies `App`.

## Defining a command

Define a config struct and a command struct that embeds `BaseCommand[T]`:

{{ file "docs/templates/inputs/sdk_defining_command_example.md" }}

`BaseCommand[T]` provides `Name`, `Add`, `Options`, `Commands`, and `Validate` for free. You implement `Run` and `Help`.

Parsed values are available on `c.Inputs` after the framework populates the struct from CLI args, env vars, and defaults.

## Registering commands

{{ file "docs/templates/inputs/sdk_registering_commands_example.md" }}

`Add` returns the command, so you can chain subcommand registration:

{{ file "docs/templates/inputs/sdk_registering_subcommands_example.md" }}

## Default command

A default command runs when the binary is invoked with no arguments. It receives values from environment variables only (no CLI flags):

{{ file "docs/templates/inputs/sdk_default_command_example.md" }}

## Positional arguments

Use numeric `arg` tags for positional args:

{{ file "docs/templates/inputs/sdk_positional_args_example.md" }}

## Dotenv

Load `.env` files before command dispatch:

{{ file "docs/templates/inputs/sdk_dotenv_example.md" }}

Missing files are silently skipped. Variables already set in the environment are never overwritten.

## Providing examples

Implement `ExampleProvider` to add usage examples to help output:

{{ file "docs/templates/inputs/sdk_examples_provider_example.md" }}

## Unknown args and options

Commands receive an `Unknowns` struct with arguments and flags that were not matched to any defined field. This supports pass-through patterns:

{{ file "docs/templates/inputs/sdk_unknowns_passthrough_example.md" }}

## Built-in commands

The module ships ready-made commands:

{{ file "docs/templates/inputs/sdk_builtin_commands_registration_example.md" }}

{{ file "docs/templates/inputs/sdk_builtin_commands_description_table.md" }}

## Parent placeholders

Group subcommands under a namespace without a handler:

{{ file "docs/templates/inputs/sdk_parent_placeholders_example.md" }}

Running `myapp db` without a subcommand displays the available subcommands.

## Config store

File-based key-value storage:

{{ file "docs/templates/inputs/sdk_config_store_store_interface.md" }}

{{ file "docs/templates/inputs/sdk_config_store_codec_interface.md" }}

Create a store:

{{ file "docs/templates/inputs/sdk_config_store_create_example.md" }}

Use it:

{{ file "docs/templates/inputs/sdk_config_store_usage_example.md" }}

JSON is built in. Register additional codecs for YAML or TOML:

{{ file "docs/templates/inputs/sdk_config_store_codecs_example.md" }}

## Config discovery

Walk up the directory tree to find config files:

{{ file "docs/templates/inputs/sdk_config_discovery_example.md" }}

## Running the app

{{ file "docs/templates/inputs/sdk_running_app_example.md" }}

Sentinel errors returned by `Run`:

{{ file "docs/templates/inputs/sdk_running_errors_table.md" }}

## Complete example

{{ file "docs/templates/inputs/sdk_complete_example.md" }}
