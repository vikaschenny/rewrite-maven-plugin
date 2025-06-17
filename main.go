package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Global configuration
	config *Config

	// Command line flags
	configFile    string
	activeRecipes []string
	activeStyles  []string
	baseDir       string
	dryRun        bool
	skip          bool
	verbose       bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rewrite-go",
	Short: "OpenRewrite code transformation tool in Go",
	Long: `A Go implementation of the OpenRewrite Maven plugin that applies code transformation recipes to your projects.

This tool provides similar functionality to the Maven rewrite plugin:
- Apply transformation recipes to source code
- Support for multiple file types and languages
- Configuration via YAML files
- Dry-run capability to preview changes

Examples:
  rewrite-go run                                    # Run with default configuration
  rewrite-go run --config custom-rewrite.yml       # Use custom config file
  rewrite-go run --active-recipes Recipe1,Recipe2  # Specify recipes
  rewrite-go dry-run                               # Preview changes without applying
  rewrite-go discover                              # List available recipes`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig()
	},
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run recipes and apply changes",
	Long: `Run the configured recipes and apply changes to source files.

This command will:
1. Load the configuration and recipes
2. Discover source files in the project
3. Apply active recipes to the files
4. Write the transformed files back to disk

Use --dry-run to preview changes without applying them.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRewrite(false)
	},
}

// dryRunCmd represents the dry-run command
var dryRunCmd = &cobra.Command{
	Use:   "dry-run",
	Short: "Preview changes without applying them",
	Long: `Run recipes and show what changes would be made without actually applying them.

This is useful for:
- Previewing transformations before applying them
- Understanding what recipes would do to your code
- Testing recipe configurations`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRewrite(true)
	},
}

// discoverCmd represents the discover command
var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "List available recipes",
	Long: `Discover and list all available recipes that can be applied.

This command will scan the classpath and configuration files to find
all available transformation recipes and display them with their descriptions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return discoverRecipes()
	},
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(dryRunCmd)
	rootCmd.AddCommand(discoverCmd)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is rewrite.yml)")
	rootCmd.PersistentFlags().StringSliceVar(&activeRecipes, "active-recipes", []string{}, "comma-separated list of recipes to activate")
	rootCmd.PersistentFlags().StringSliceVar(&activeStyles, "active-styles", []string{}, "comma-separated list of styles to activate")
	rootCmd.PersistentFlags().StringVar(&baseDir, "base-dir", "", "base directory to process (default is current directory)")
	rootCmd.PersistentFlags().BoolVar(&skip, "skip", false, "skip execution")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Command-specific flags
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview changes without applying them")

	// Bind flags to viper
	viper.BindPFlag("config-location", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("active-recipes", rootCmd.PersistentFlags().Lookup("active-recipes"))
	viper.BindPFlag("active-styles", rootCmd.PersistentFlags().Lookup("active-styles"))
	viper.BindPFlag("skip", rootCmd.PersistentFlags().Lookup("skip"))
	viper.BindPFlag("dry-run", runCmd.Flags().Lookup("dry-run"))
}

// initConfig reads in config file and ENV variables if set
func initConfig() error {
	// Create default configuration
	config = NewDefaultConfig()

	// Set base directory
	if baseDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		baseDir = cwd
	}

	// Use config file from the flag
	if configFile != "" {
		viper.SetConfigFile(configFile)
		config.ConfigLocation = configFile
	} else {
		// Search for config in current directory and home directory
		viper.SetConfigName("rewrite")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")

		// Set default config location
		config.ConfigLocation = "rewrite.yml"
	}

	// Read environment variables
	viper.SetEnvPrefix("REWRITE")
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is okay, we'll use defaults
	}

	// Unmarshal configuration
	err := viper.Unmarshal(config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override with command line flags
	if len(activeRecipes) > 0 {
		config.ActiveRecipes = activeRecipes
	}
	if len(activeStyles) > 0 {
		config.ActiveStyles = activeStyles
	}
	if skip {
		config.Skip = true
	}

	// Set log level based on verbose flag
	if verbose {
		config.LogLevel = "debug"
	}

	return nil
}

// runRewrite executes the rewrite operation
func runRewrite(isDryRun bool) error {
	// Validate config
	if config == nil {
		return fmt.Errorf("configuration not initialized")
	}

	// Create rewriter
	rewriter := NewRewriter(config, baseDir)

	// Create runner
	runner := NewRunner(rewriter)

	// Execute
	if isDryRun || dryRun {
		return runner.DryRun()
	} else {
		return runner.Execute()
	}
}

// discoverRecipes lists available recipes
func discoverRecipes() error {
	fmt.Println("Available recipes:")
	fmt.Println("(This is a placeholder - recipe discovery would be implemented here)")

	// Create rewriter to load environment
	rewriter := NewRewriter(config, baseDir)
	err := rewriter.LoadEnvironment()
	if err != nil {
		return fmt.Errorf("failed to load environment: %w", err)
	}

	if rewriter.Environment != nil {
		fmt.Printf("\nLoaded %d recipes from configuration:\n", len(rewriter.Environment.ActiveRecipes))
		for _, recipe := range rewriter.Environment.ActiveRecipes {
			fmt.Printf("  - %s", recipe.Name)
			if recipe.DisplayName != "" {
				fmt.Printf(" (%s)", recipe.DisplayName)
			}
			if recipe.Description != "" {
				fmt.Printf(": %s", recipe.Description)
			}
			fmt.Println()
		}

		fmt.Printf("\nLoaded %d styles from configuration:\n", len(rewriter.Environment.ActiveStyles))
		for _, style := range rewriter.Environment.ActiveStyles {
			fmt.Printf("  - %s\n", style.Name)
		}
	}

	return nil
}

// main is the entry point
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// version information (could be set via build flags)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	// Add version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("rewrite-go version %s (commit: %s, built: %s)\n", version, commit, date)
		},
	}
	rootCmd.AddCommand(versionCmd)
}
