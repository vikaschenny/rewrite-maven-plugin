package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Rewriter represents the core rewrite engine
// This mirrors the AbstractRewriteMojo class functionality
type Rewriter struct {
	Config      *Config
	Environment *Environment
	BaseDir     string
}

// Environment represents the rewrite environment with loaded recipes and configurations
// This mirrors the Environment class from the Java version
type Environment struct {
	ActiveRecipes []Recipe
	ActiveStyles  []Style
	Properties    map[string]string
}

// Recipe represents a rewrite recipe
type Recipe struct {
	Name        string                 `yaml:"name"`
	DisplayName string                 `yaml:"displayName,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Tags        []string               `yaml:"tags,omitempty"`
	RecipeList  []string               `yaml:"recipeList,omitempty"`
	Config      map[string]interface{} `yaml:",inline"`
}

// Style represents a rewrite style configuration
type Style struct {
	Name   string                 `yaml:"name"`
	Config map[string]interface{} `yaml:",inline"`
}

// RewriteConfig represents the structure of rewrite.yml
type RewriteConfig struct {
	Type        string   `yaml:"type,omitempty"`
	Recipes     []Recipe `yaml:"recipes,omitempty"`
	Styles      []Style  `yaml:"styles,omitempty"`
	RecipeList  []string `yaml:"recipeList,omitempty"`
	StyleList   []string `yaml:"styleList,omitempty"`
	Description string   `yaml:"description,omitempty"`
}

// Result represents the result of a rewrite operation
// This mirrors the Result class from the Java version
type Result struct {
	Before                 *SourceFile
	After                  *SourceFile
	RecipesThatMadeChanges []string
	TimeSaved              time.Duration
}

// SourceFile represents a source file being processed
type SourceFile struct {
	Path     string
	Content  string
	Charset  string
	Modified bool
}

// ResultsContainer holds all the results from rewrite operations
// This mirrors the ResultsContainer from AbstractRewriteRunMojo
type ResultsContainer struct {
	Generated         []Result
	Deleted           []Result
	Moved             []Result
	RefactoredInPlace []Result
	ProjectRoot       string
	FirstException    error
}

// NewRewriter creates a new Rewriter instance
func NewRewriter(config *Config, baseDir string) *Rewriter {
	return &Rewriter{
		Config:  config,
		BaseDir: baseDir,
	}
}

// LoadEnvironment loads the rewrite environment from configuration
// This mirrors the environment() method from AbstractRewriteMojo
func (r *Rewriter) LoadEnvironment() error {
	env := &Environment{
		Properties: make(map[string]string),
	}

	// Load configuration file if it exists
	configLocation, err := r.Config.GetConfigLocation()
	if err != nil {
		return fmt.Errorf("failed to get config location: %w", err)
	}

	if configLocation != "" {
		err = r.loadConfigurationFile(configLocation, env)
		if err != nil {
			return fmt.Errorf("failed to load configuration file: %w", err)
		}
	}

	// Apply active recipes filter
	r.filterActiveRecipes(env)
	r.filterActiveStyles(env)

	r.Environment = env
	return nil
}

// loadConfigurationFile loads configuration from a file or URL
// This mirrors the getConfig() method logic from AbstractRewriteMojo
func (r *Rewriter) loadConfigurationFile(location string, env *Environment) error {
	var content []byte
	var err error

	// Check if it's a URL
	if strings.HasPrefix(location, "http") {
		resp, err := http.Get(location)
		if err != nil {
			return fmt.Errorf("failed to fetch config from URL: %w", err)
		}
		defer resp.Body.Close()

		content, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read config from URL: %w", err)
		}
	} else {
		// Load from file
		content, err = os.ReadFile(location)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Parse YAML configuration
	var rewriteConfig RewriteConfig
	err = yaml.Unmarshal(content, &rewriteConfig)
	if err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Load recipes and styles into environment
	env.ActiveRecipes = append(env.ActiveRecipes, rewriteConfig.Recipes...)
	env.ActiveStyles = append(env.ActiveStyles, rewriteConfig.Styles...)

	// Add recipes from recipeList
	for _, recipeName := range rewriteConfig.RecipeList {
		env.ActiveRecipes = append(env.ActiveRecipes, Recipe{Name: recipeName})
	}

	// Add styles from styleList
	for _, styleName := range rewriteConfig.StyleList {
		env.ActiveStyles = append(env.ActiveStyles, Style{Name: styleName})
	}

	return nil
}

// filterActiveRecipes filters recipes based on configuration
func (r *Rewriter) filterActiveRecipes(env *Environment) {
	activeRecipeNames := r.Config.GetActiveRecipes()
	if len(activeRecipeNames) == 0 {
		return
	}

	nameSet := make(map[string]bool)
	for _, name := range activeRecipeNames {
		nameSet[name] = true
	}

	var filteredRecipes []Recipe
	for _, recipe := range env.ActiveRecipes {
		if nameSet[recipe.Name] {
			filteredRecipes = append(filteredRecipes, recipe)
		}
	}

	env.ActiveRecipes = filteredRecipes
}

// filterActiveStyles filters styles based on configuration
func (r *Rewriter) filterActiveStyles(env *Environment) {
	activeStyleNames := r.Config.GetActiveStyles()
	if len(activeStyleNames) == 0 {
		return
	}

	nameSet := make(map[string]bool)
	for _, name := range activeStyleNames {
		nameSet[name] = true
	}

	var filteredStyles []Style
	for _, style := range env.ActiveStyles {
		if nameSet[style.Name] {
			filteredStyles = append(filteredStyles, style)
		}
	}

	env.ActiveStyles = filteredStyles
}

// GetBuildRoot determines the root directory for the build
// This mirrors the getBuildRoot() method from AbstractRewriteMojo
func (r *Rewriter) GetBuildRoot() (string, error) {
	if r.BaseDir != "" {
		abs, err := filepath.Abs(r.BaseDir)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path: %w", err)
		}
		return abs, nil
	}

	// Default to current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	return cwd, nil
}

// FindSourceFiles discovers source files to process
func (r *Rewriter) FindSourceFiles(rootDir string) ([]string, error) {
	var sourceFiles []string
	exclusions := r.Config.GetExclusions()
	plainTextMasks := r.Config.GetPlainTextMasks()

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check file size threshold
		sizeMB := float64(info.Size()) / (1024 * 1024)
		if sizeMB > float64(r.Config.SizeThresholdMb) {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		// Check exclusions
		if r.matchesPatterns(relPath, exclusions) {
			return nil
		}

		// Check if it matches plain text masks or is a known source file type
		if r.matchesPatterns(relPath, plainTextMasks) || r.isSourceFile(relPath) {
			sourceFiles = append(sourceFiles, path)
		}

		return nil
	})

	return sourceFiles, err
}

// matchesPatterns checks if a path matches any of the given patterns
func (r *Rewriter) matchesPatterns(path string, patterns []string) bool {
	for _, pattern := range patterns {
		// Simple glob pattern matching (this is a simplified implementation)
		// In a real implementation, you'd want to use a proper glob library
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}

		// Handle ** patterns
		if strings.Contains(pattern, "**") {
			parts := strings.Split(pattern, "**")
			if len(parts) == 2 {
				prefix := parts[0]
				suffix := parts[1]
				if strings.HasPrefix(path, prefix) && strings.HasSuffix(path, suffix) {
					return true
				}
			}
		}
	}
	return false
}

// isSourceFile determines if a file is a source file based on extension
func (r *Rewriter) isSourceFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	sourceExtensions := []string{
		".java", ".kt", ".groovy", ".scala",
		".js", ".ts", ".jsx", ".tsx",
		".go", ".rs", ".py", ".rb",
		".c", ".cpp", ".h", ".hpp",
		".cs", ".vb", ".php",
		".xml", ".json", ".yaml", ".yml",
		".properties", ".toml", ".hcl",
	}

	for _, sourceExt := range sourceExtensions {
		if ext == sourceExt {
			return true
		}
	}

	return false
}

// ProcessFiles applies recipes to the discovered source files
func (r *Rewriter) ProcessFiles(sourceFiles []string) (*ResultsContainer, error) {
	if r.Environment == nil {
		return nil, fmt.Errorf("environment not loaded")
	}

	results := &ResultsContainer{
		ProjectRoot: r.BaseDir,
	}

	for _, filePath := range sourceFiles {
		result, err := r.processFile(filePath)
		if err != nil {
			if results.FirstException == nil {
				results.FirstException = err
			}
			continue
		}

		if result != nil {
			// Categorize the result
			if result.Before == nil {
				results.Generated = append(results.Generated, *result)
			} else if result.After == nil {
				results.Deleted = append(results.Deleted, *result)
			} else if result.Before.Path != result.After.Path {
				results.Moved = append(results.Moved, *result)
			} else if result.Before.Content != result.After.Content {
				results.RefactoredInPlace = append(results.RefactoredInPlace, *result)
			}
		}
	}

	return results, nil
}

// processFile processes a single file through the active recipes
func (r *Rewriter) processFile(filePath string) (*Result, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	relPath, err := filepath.Rel(r.BaseDir, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}

	before := &SourceFile{
		Path:     relPath,
		Content:  string(content),
		Charset:  "UTF-8",
		Modified: false,
	}

	// Apply recipes (this is a simplified placeholder)
	// In a real implementation, this would invoke the actual OpenRewrite recipes
	after := r.applyRecipes(before)

	if before.Content == after.Content {
		return nil, nil // No changes
	}

	return &Result{
		Before:                 before,
		After:                  after,
		RecipesThatMadeChanges: r.getActiveRecipeNames(),
		TimeSaved:              time.Minute, // Placeholder
	}, nil
}

// applyRecipes applies the active recipes to a source file
// This is a simplified placeholder implementation
func (r *Rewriter) applyRecipes(sourceFile *SourceFile) *SourceFile {
	// This is where the actual recipe application would happen
	// For now, this is a placeholder that doesn't modify anything
	after := &SourceFile{
		Path:     sourceFile.Path,
		Content:  sourceFile.Content,
		Charset:  sourceFile.Charset,
		Modified: false,
	}

	// TODO: Implement actual recipe application logic
	// This would involve:
	// 1. Parsing the source file into an AST
	// 2. Applying each active recipe to the AST
	// 3. Converting the modified AST back to source code

	return after
}

// getActiveRecipeNames returns the names of active recipes
func (r *Rewriter) getActiveRecipeNames() []string {
	var names []string
	for _, recipe := range r.Environment.ActiveRecipes {
		names = append(names, recipe.Name)
	}
	return names
}

// IsNotEmpty checks if the results container has any results
func (rc *ResultsContainer) IsNotEmpty() bool {
	return len(rc.Generated) > 0 || len(rc.Deleted) > 0 ||
		len(rc.Moved) > 0 || len(rc.RefactoredInPlace) > 0
}
