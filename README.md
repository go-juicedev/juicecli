# JuiceCLI

JuiceCLI is a powerful command-line tool designed to enhance development efficiency with the Juice ORM framework.

## Features

- **Interface Implementation Generator**: Automatically generate implementation code for interfaces
- **Smart Namespace Suggestion**: Get intelligent suggestions for package namespaces
- **XML Configuration**: Customize implementation details through XML configuration

## Installation

```bash
go install github.com/go-juicedev/juicecli@latest
```

## Usage

### Generate Implementation

Generate implementation for an interface:

```bash
juicecli impl --type UserRepository
```

Options:
- `--type, -t`: The interface type name to generate implementation for (required)
- `--namespace, -n`: The package name for the generated implementation. If not specified, it will be auto-generated
- `--output, -o`: The output file path. If not specified, output will be written to stdout
- `--config, -c`: The configuration file path. If not specified, it will search for:
  - juice.xml
  - config/juice.xml
  - config.xml
  - config/config.xml

Examples:
```bash
# Basic usage
juicecli impl --type UserRepository

# With custom namespace and output file
juicecli impl --type UserRepository --namespace repository --output user_repository.go

# With custom config file
juicecli impl --type UserRepository --config custom.xml
```

### Get Namespace Suggestion

Get a suggested namespace for your interface:

```bash
juicecli tell --type UserRepository
```

Options:
- `--type, -t`: The interface type name to analyze (required)

## Configuration

The implementation generator can be customized through XML configuration files. Example configuration:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<configuration>
    <settings>
        <!-- Custom settings here -->
    </settings>
</configuration>
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
