package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Config represents the configuration for the rewrite tool
// This mirrors the ConfigurableRewriteMojo class from the Java version
type Config struct {
	// ConfigLocation is the path to rewrite.yml configuration file
	ConfigLocation string `yaml:"configLocation" mapstructure:"config-location"`

	// ActiveRecipes is the list of recipes to activate
	ActiveRecipes []string `yaml:"activeRecipes" mapstructure:"active-recipes"`

	// ActiveStyles is the list of styles to activate
	ActiveStyles []string `yaml:"activeStyles" mapstructure:"active-styles"`

	// PomCacheEnabled determines if POM caching is enabled
	PomCacheEnabled bool `yaml:"pomCacheEnabled" mapstructure:"pom-cache-enabled"`

	// PomCacheDirectory is the directory for POM cache
	PomCacheDirectory string `yaml:"pomCacheDirectory" mapstructure:"pom-cache-directory"`

	// Skip determines if rewrite execution should be skipped
	Skip bool `yaml:"skip" mapstructure:"skip"`

	// SkipMavenParsing skips parsing Maven pom.xml files
	SkipMavenParsing bool `yaml:"skipMavenParsing" mapstructure:"skip-maven-parsing"`

	// CheckstyleConfigFile is the path to checkstyle configuration
	CheckstyleConfigFile string `yaml:"checkstyleConfigFile" mapstructure:"checkstyle-config-file"`

	// CheckstyleDetectionEnabled enables automatic checkstyle detection
	CheckstyleDetectionEnabled bool `yaml:"checkstyleDetectionEnabled" mapstructure:"checkstyle-detection-enabled"`

	// Exclusions are file patterns to exclude from processing
	Exclusions []string `yaml:"exclusions" mapstructure:"exclusions"`

	// PlainTextMasks are patterns for plain text files
	PlainTextMasks []string `yaml:"plainTextMasks" mapstructure:"plain-text-masks"`

	// AdditionalPlainTextMasks are additional patterns for plain text files
	AdditionalPlainTextMasks []string `yaml:"additionalPlainTextMasks" mapstructure:"additional-plain-text-masks"`

	// SizeThresholdMb is the size threshold in MB for processing files
	SizeThresholdMb int `yaml:"sizeThresholdMb" mapstructure:"size-threshold-mb"`

	// FailOnInvalidActiveRecipes determines if invalid recipes should fail the execution
	FailOnInvalidActiveRecipes bool `yaml:"failOnInvalidActiveRecipes" mapstructure:"fail-on-invalid-active-recipes"`

	// RunPerSubmodule determines if execution should run per submodule
	RunPerSubmodule bool `yaml:"runPerSubmodule" mapstructure:"run-per-submodule"`

	// RecipeArtifactCoordinates are Maven coordinates for recipe artifacts
	RecipeArtifactCoordinates []string `yaml:"recipeArtifactCoordinates" mapstructure:"recipe-artifact-coordinates"`

	// ResolvePropertiesInYaml determines if properties should be resolved in YAML
	ResolvePropertiesInYaml bool `yaml:"resolvePropertiesInYaml" mapstructure:"resolve-properties-in-yaml"`

	// LogLevel for the rewrite execution
	LogLevel string `yaml:"logLevel" mapstructure:"log-level"`

	// ExportDatatables determines if datatables should be exported
	ExportDatatables bool `yaml:"exportDatatables" mapstructure:"export-datatables"`
}

// NewDefaultConfig creates a new Config with default values
// This mirrors the default values from the Java Maven plugin
func NewDefaultConfig() *Config {
	return &Config{
		ConfigLocation:             "rewrite.yml",
		PomCacheEnabled:            true,
		CheckstyleDetectionEnabled: true,
		SizeThresholdMb:            10,
		FailOnInvalidActiveRecipes: false,
		RunPerSubmodule:            false,
		ResolvePropertiesInYaml:    true,
		LogLevel:                   "info",
		ExportDatatables:           false,
		PlainTextMasks:             getDefaultPlainTextMasks(),
	}
}

// getDefaultPlainTextMasks returns the default plain text file patterns
// This mirrors the default masks from ConfigurableRewriteMojo
func getDefaultPlainTextMasks() []string {
	return []string{
		"**/*.adoc",
		"**/*.aj",
		"**/*.bash",
		"**/*.bat",
		"**/CODEOWNERS",
		"**/*.css",
		"**/*.config",
		"**/[dD]ockerfile*",
		"**/*.[dD]ockerfile",
		"**/*[cC]ontainerfile*",
		"**/*.[cC]ontainerfile",
		"**/*.env",
		"**/.gitattributes",
		"**/.gitignore",
		"**/*.htm*",
		"**/gradlew",
		"**/.java-version",
		"**/*.jelly",
		"**/*.jsp",
		"**/*.ksh",
		"**/*.lock",
		"**/lombok.config",
		"**/*.md",
		"**/*.mf",
		"**/META-INF/services/**",
		"**/META-INF/spring/**",
		"**/META-INF/spring.factories",
		"**/mvnw",
		"**/mvnw.cmd",
		"**/*.qute.java",
		"**/.sdkmanrc",
		"**/*.sh",
		"**/*.sql",
		"**/*.svg",
		"**/*.tsx",
		"**/*.txt",
		"**/*.py",
	}
}

// GetConfigLocation resolves the configuration file location
// This mirrors the getConfig() method from AbstractRewriteMojo
func (c *Config) GetConfigLocation() (string, error) {
	if c.ConfigLocation == "" {
		return "", nil
	}

	// Check if it's a URL
	if u, err := url.Parse(c.ConfigLocation); err == nil && u.Scheme != "" && strings.HasPrefix(u.Scheme, "http") {
		return c.ConfigLocation, nil
	}

	// Treat as file path
	configPath := c.ConfigLocation
	if !filepath.IsAbs(configPath) {
		// Make it relative to current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
		configPath = filepath.Join(cwd, configPath)
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", nil // File doesn't exist, return empty string
	} else if err != nil {
		return "", fmt.Errorf("failed to check config file: %w", err)
	}

	return configPath, nil
}

// GetPlainTextMasks returns the effective plain text masks
// This mirrors the getPlainTextMasks() method logic
func (c *Config) GetPlainTextMasks() []string {
	if len(c.PlainTextMasks) > 0 {
		return c.PlainTextMasks
	}

	// Use defaults and add additional masks
	masks := make([]string, len(getDefaultPlainTextMasks()))
	copy(masks, getDefaultPlainTextMasks())
	masks = append(masks, c.AdditionalPlainTextMasks...)

	return masks
}

// CleanStringSlice removes empty and whitespace-only strings from a slice
// This mirrors the getCleanedSet() method from ConfigurableRewriteMojo
func CleanStringSlice(input []string) []string {
	var cleaned []string
	seen := make(map[string]bool)

	for _, s := range input {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" && !seen[trimmed] {
			cleaned = append(cleaned, trimmed)
			seen[trimmed] = true
		}
	}

	return cleaned
}

// GetActiveRecipes returns cleaned active recipes
func (c *Config) GetActiveRecipes() []string {
	return CleanStringSlice(c.ActiveRecipes)
}

// GetActiveStyles returns cleaned active styles
func (c *Config) GetActiveStyles() []string {
	return CleanStringSlice(c.ActiveStyles)
}

// GetExclusions returns cleaned exclusions
func (c *Config) GetExclusions() []string {
	return CleanStringSlice(c.Exclusions)
}

// GetRecipeArtifactCoordinates returns cleaned recipe artifact coordinates
func (c *Config) GetRecipeArtifactCoordinates() []string {
	return CleanStringSlice(c.RecipeArtifactCoordinates)
}
