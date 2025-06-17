# Rewrite-Go

A Go implementation of the OpenRewrite Maven plugin that applies code transformation recipes to your projects.

This tool provides similar functionality to the Maven rewrite plugin, allowing you to apply various code transformation recipes to source code written in multiple languages.

## Conversion from Java Maven Plugin

This Go version mirrors the structure and functionality of the original Java Maven plugin:

- **ConfigurableRewriteMojo** → `config.go` - Configuration management and parameter handling
- **AbstractRewriteMojo** → `rewriter.go` - Core rewrite engine and environment setup
- **AbstractRewriteRunMojo** → `runner.go` - Execution logic and file operations
- **Maven Plugin Annotations** → `main.go` - CLI interface using Cobra

## Features

- ✅ Configuration via YAML files
- ✅ Command-line interface with multiple subcommands
- ✅ Dry-run capability to preview changes
- ✅ Support for multiple file types and patterns
- ✅ Recipe and style management
- ✅ File exclusion patterns
- ✅ Size threshold filtering
- ✅ Environment variable support
- ⚠️ Recipe application logic (placeholder - needs actual OpenRewrite recipe implementation)

## Installation

```bash
# Clone the repository
git clone https://github.com/openrewrite/rewrite-go
cd rewrite-go

# Build the binary
go build -o rewrite-go

# Install globally (optional)
go install
```

## Usage

### Basic Commands

```bash
# Run recipes and apply changes
./rewrite-go run

# Preview changes without applying (dry run)
./rewrite-go dry-run

# List available recipes
./rewrite-go discover

# Show help
./rewrite-go --help

# Show version
./rewrite-go version
```

### Configuration Options

```bash
# Use a custom configuration file
./rewrite-go run --config custom-rewrite.yml

# Specify active recipes via command line
./rewrite-go run --active-recipes "Recipe1,Recipe2,Recipe3"

# Specify active styles
./rewrite-go run --active-styles "Style1,Style2"

# Run on a specific directory
./rewrite-go run --base-dir /path/to/project

# Verbose output
./rewrite-go run --verbose

# Skip execution
./rewrite-go run --skip
```

### Configuration File

Create a `rewrite.yml` file in your project root:

```yaml
# Example rewrite.yml configuration
type: specs.openrewrite.org/v1beta/recipe
name: example.MyRecipe
displayName: My Custom Recipe
description: An example recipe configuration

# Recipe list - recipes to activate
recipeList:
  - org.openrewrite.java.format.AutoFormat
  - org.openrewrite.java.RemoveUnusedImports
  - org.openrewrite.java.OrderImports

# Style list - styles to activate  
styleList:
  - org.openrewrite.java.IntelliJ

# Recipe definitions
recipes:
  - name: example.CustomRule
    displayName: Custom Rule
    description: A custom transformation rule
    # Recipe-specific configuration would go here

# Style definitions
styles:
  - name: example.CustomStyle
    # Style-specific configuration would go here

# Global configuration
excludes:
  - "**/target/**"
  - "**/build/**"
  - "**/.git/**"

plainTextMasks:
  - "**/*.txt"
  - "**/*.md"
  - "**/*.json"

sizeThresholdMb: 10
pomCacheEnabled: true
checkstyleDetectionEnabled: true
```

### Environment Variables

You can configure the tool using environment variables with the `REWRITE_` prefix:

```bash
export REWRITE_CONFIG_LOCATION=custom-config.yml
export REWRITE_ACTIVE_RECIPES=Recipe1,Recipe2
export REWRITE_SKIP=true
export REWRITE_LOG_LEVEL=debug

./rewrite-go run
```

## File Support

The tool automatically detects and processes various file types:

**Source Code Files:**
- Java (`.java`)
- Kotlin (`.kt`)
- Groovy (`.groovy`)
- Scala (`.scala`)
- JavaScript/TypeScript (`.js`, `.ts`, `.jsx`, `.tsx`)
- Go (`.go`)
- Rust (`.rs`)
- Python (`.py`)
- Ruby (`.rb`)
- C/C++ (`.c`, `.cpp`, `.h`, `.hpp`)
- C# (`.cs`)
- PHP (`.php`)

**Configuration Files:**
- XML (`.xml`)
- JSON (`.json`)
- YAML (`.yaml`, `.yml`)
- Properties (`.properties`)
- TOML (`.toml`)
- HCL (`.hcl`)

**Plain Text Files:**
- Markdown (`.md`)
- Text (`.txt`)
- Shell scripts (`.sh`, `.bash`)
- Batch files (`.bat`)
- Dockerfiles
- And many more (see `config.go` for full list)

## Architecture

The Go version follows the same architectural patterns as the Java Maven plugin:

### Core Components

1. **Config (`config.go`)** - Configuration management
   - YAML configuration loading
   - Command-line parameter handling
   - Environment variable support
   - Default value management

2. **Rewriter (`rewriter.go`)** - Core transformation engine
   - Environment setup and recipe loading
   - File discovery and filtering
   - Recipe application (placeholder for actual OpenRewrite integration)
   - Result processing

3. **Runner (`runner.go`)** - Execution management
   - Dry-run vs. actual execution
   - File writing and backup operations
   - Progress reporting and logging
   - Error handling

4. **Main (`main.go`)** - CLI interface
   - Command-line parsing using Cobra
   - Configuration initialization
   - Subcommand routing

### Maven Plugin Equivalents

| Java Class | Go File | Purpose |
|------------|---------|---------|
| `ConfigurableRewriteMojo` | `config.go` | Configuration and parameters |
| `AbstractRewriteMojo` | `rewriter.go` | Core rewrite functionality |
| `AbstractRewriteRunMojo` | `runner.go` | Execution logic |
| `RewriteRunMojo` | CLI run command | Execute recipes |
| `RewriteDryRunMojo` | CLI dry-run command | Preview changes |
| `RewriteDiscoverMojo` | CLI discover command | List recipes |

## Development Status

This is a conversion/port of the Java Maven plugin to Go. The structure and interfaces are complete, but the actual recipe application logic is currently a placeholder.

### What's Implemented ✅

- Complete CLI interface with all major commands
- Configuration file loading and management
- File discovery and filtering
- Dry-run functionality
- Result categorization and reporting
- File writing and directory cleanup
- Environment variable support
- Error handling and logging

### What Needs Implementation ⚠️

- Actual OpenRewrite recipe execution engine
- AST parsing and transformation for different languages
- Integration with OpenRewrite recipe ecosystem
- Recipe artifact resolution and loading
- Checkstyle configuration parsing
- Maven POM parsing (if needed)

## Contributing

This is a faithful port of the Java Maven plugin. When implementing new features or fixing bugs, please refer to the original Java implementation to maintain compatibility.

### Key Design Principles

1. **Maintain API Compatibility** - Configuration options and behavior should match the Maven plugin
2. **Preserve Structure** - The Go code mirrors the Java class hierarchy and responsibilities
3. **Idiomatic Go** - While maintaining compatibility, use Go best practices and idioms
4. **Cross-Platform** - Ensure the tool works on Windows, macOS, and Linux

## License

This project follows the same license as the original OpenRewrite Maven plugin.

## Links

- [Original OpenRewrite Maven Plugin](https://github.com/openrewrite/rewrite-maven-plugin)
- [OpenRewrite Documentation](https://docs.openrewrite.org/)
- [OpenRewrite Recipes](https://docs.openrewrite.org/recipes)
