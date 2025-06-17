package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Runner handles the execution of rewrite operations
// This mirrors the AbstractRewriteRunMojo functionality
type Runner struct {
	Rewriter *Rewriter
	Logger   *log.Logger
}

// NewRunner creates a new Runner instance
func NewRunner(rewriter *Rewriter) *Runner {
	return &Runner{
		Rewriter: rewriter,
		Logger:   log.New(os.Stdout, "[REWRITE] ", log.LstdFlags),
	}
}

// Execute runs the rewrite operation
// This mirrors the execute() method from AbstractRewriteRunMojo
func (r *Runner) Execute() error {
	if r.Rewriter.Config.Skip {
		r.Logger.Println("Skipping execution")
		return nil
	}

	// Load the environment
	err := r.Rewriter.LoadEnvironment()
	if err != nil {
		return fmt.Errorf("failed to load environment: %w", err)
	}

	// Get the build root
	buildRoot, err := r.Rewriter.GetBuildRoot()
	if err != nil {
		return fmt.Errorf("failed to get build root: %w", err)
	}

	r.Logger.Printf("Processing project at: %s", buildRoot)

	// Find source files
	sourceFiles, err := r.Rewriter.FindSourceFiles(buildRoot)
	if err != nil {
		return fmt.Errorf("failed to find source files: %w", err)
	}

	r.Logger.Printf("Found %d source files to process", len(sourceFiles))

	if len(sourceFiles) == 0 {
		r.Logger.Println("No source files found to process")
		return nil
	}

	// Process the files
	results, err := r.Rewriter.ProcessFiles(sourceFiles)
	if err != nil {
		return fmt.Errorf("failed to process files: %w", err)
	}

	// Handle first exception if any
	if results.FirstException != nil {
		r.Logger.Printf("ERROR: The recipe produced an error: %v", results.FirstException)
		return results.FirstException
	}

	// Report results
	if results.IsNotEmpty() {
		err = r.reportAndApplyResults(results)
		if err != nil {
			return fmt.Errorf("failed to apply results: %w", err)
		}
	} else {
		r.Logger.Println("No changes were made")
	}

	return nil
}

// reportAndApplyResults reports the results and applies the changes
// This mirrors the result processing logic from AbstractRewriteRunMojo
func (r *Runner) reportAndApplyResults(results *ResultsContainer) error {
	var totalTimeSaved time.Duration

	// Report generated files
	for _, result := range results.Generated {
		if result.After != nil {
			r.Logger.Printf("Generated new file %s by:", result.After.Path)
			r.logRecipesThatMadeChanges(result.RecipesThatMadeChanges)
			totalTimeSaved += result.TimeSaved
		}
	}

	// Report deleted files
	for _, result := range results.Deleted {
		if result.Before != nil {
			r.Logger.Printf("Deleted file %s by:", result.Before.Path)
			r.logRecipesThatMadeChanges(result.RecipesThatMadeChanges)
			totalTimeSaved += result.TimeSaved
		}
	}

	// Report moved files
	for _, result := range results.Moved {
		if result.Before != nil && result.After != nil {
			r.Logger.Printf("File has been moved from %s to %s by:", result.Before.Path, result.After.Path)
			r.logRecipesThatMadeChanges(result.RecipesThatMadeChanges)
			totalTimeSaved += result.TimeSaved
		}
	}

	// Report refactored files
	for _, result := range results.RefactoredInPlace {
		if result.Before != nil {
			r.Logger.Printf("Changes have been made to %s by:", result.Before.Path)
			r.logRecipesThatMadeChanges(result.RecipesThatMadeChanges)
			totalTimeSaved += result.TimeSaved
		}
	}

	r.Logger.Println("Please review and commit the results.")
	r.Logger.Printf("Estimate time saved: %s", r.formatDuration(totalTimeSaved))

	// Apply the changes
	err := r.applyChanges(results)
	if err != nil {
		return fmt.Errorf("failed to apply changes: %w", err)
	}

	return nil
}

// logRecipesThatMadeChanges logs the recipes that made changes
func (r *Runner) logRecipesThatMadeChanges(recipeNames []string) {
	for _, recipeName := range recipeNames {
		r.Logger.Printf("  %s", recipeName)
	}
}

// formatDuration formats a duration in a human-readable format
func (r *Runner) formatDuration(d time.Duration) string {
	if d < time.Second {
		return "< 1 second"
	}
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	return fmt.Sprintf("%.1f hours", d.Hours())
}

// applyChanges applies all the changes to the file system
// This mirrors the file writing logic from AbstractRewriteRunMojo
func (r *Runner) applyChanges(results *ResultsContainer) error {
	buildRoot := results.ProjectRoot

	// Handle generated files
	for _, result := range results.Generated {
		if result.After != nil {
			err := r.writeFile(buildRoot, result.After)
			if err != nil {
				return fmt.Errorf("failed to write generated file %s: %w", result.After.Path, err)
			}
		}
	}

	// Handle deleted files
	for _, result := range results.Deleted {
		if result.Before != nil {
			filePath := filepath.Join(buildRoot, result.Before.Path)
			err := os.Remove(filePath)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to delete file %s: %w", filePath, err)
			}
		}
	}

	// Handle moved files
	for _, result := range results.Moved {
		if result.Before != nil && result.After != nil {
			oldPath := filepath.Join(buildRoot, result.Before.Path)
			newPath := filepath.Join(buildRoot, result.After.Path)

			// Create target directory if it doesn't exist
			targetDir := filepath.Dir(newPath)
			err := os.MkdirAll(targetDir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
			}

			// Move/rename the file
			err = os.Rename(oldPath, newPath)
			if err != nil {
				// If rename fails, copy and delete
				err = r.writeFile(buildRoot, result.After)
				if err != nil {
					return fmt.Errorf("failed to write moved file %s: %w", result.After.Path, err)
				}
				err = os.Remove(oldPath)
				if err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove old file %s: %w", oldPath, err)
				}
			}
		}
	}

	// Handle refactored files
	for _, result := range results.RefactoredInPlace {
		if result.After != nil {
			err := r.writeFile(buildRoot, result.After)
			if err != nil {
				return fmt.Errorf("failed to write refactored file %s: %w", result.After.Path, err)
			}
		}
	}

	// Clean up empty directories
	err := r.cleanupEmptyDirectories(buildRoot, results)
	if err != nil {
		r.Logger.Printf("Warning: failed to cleanup empty directories: %v", err)
	}

	return nil
}

// writeFile writes a source file to disk
func (r *Runner) writeFile(buildRoot string, sourceFile *SourceFile) error {
	filePath := filepath.Join(buildRoot, sourceFile.Path)

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file
	err = os.WriteFile(filePath, []byte(sourceFile.Content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// cleanupEmptyDirectories removes directories that have become empty
func (r *Runner) cleanupEmptyDirectories(buildRoot string, results *ResultsContainer) error {
	// Collect directories that might be empty
	var candidateDirs []string

	for _, result := range results.Deleted {
		if result.Before != nil {
			dir := filepath.Dir(filepath.Join(buildRoot, result.Before.Path))
			candidateDirs = append(candidateDirs, dir)
		}
	}

	for _, result := range results.Moved {
		if result.Before != nil {
			dir := filepath.Dir(filepath.Join(buildRoot, result.Before.Path))
			candidateDirs = append(candidateDirs, dir)
		}
	}

	// Remove duplicates and sort by depth (deepest first)
	dirMap := make(map[string]bool)
	for _, dir := range candidateDirs {
		dirMap[dir] = true
	}

	var sortedDirs []string
	for dir := range dirMap {
		sortedDirs = append(sortedDirs, dir)
	}

	// Sort by depth (number of path separators), deepest first
	for i := 0; i < len(sortedDirs); i++ {
		for j := i + 1; j < len(sortedDirs); j++ {
			iDepth := strings.Count(sortedDirs[i], string(filepath.Separator))
			jDepth := strings.Count(sortedDirs[j], string(filepath.Separator))
			if iDepth < jDepth {
				sortedDirs[i], sortedDirs[j] = sortedDirs[j], sortedDirs[i]
			}
		}
	}

	// Try to remove empty directories
	var removedDirs []string
	for _, dir := range sortedDirs {
		if r.isEmptyDir(dir) {
			err := os.Remove(dir)
			if err == nil {
				removedDirs = append(removedDirs, dir)
			}
		}
	}

	if len(removedDirs) > 0 {
		r.Logger.Printf("Removed %d empty directories:", len(removedDirs))
		for _, dir := range removedDirs {
			relDir, _ := filepath.Rel(buildRoot, dir)
			r.Logger.Printf("  %s", relDir)
		}
	}

	return nil
}

// isEmptyDir checks if a directory is empty
func (r *Runner) isEmptyDir(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return len(entries) == 0
}

// DryRun performs a dry run without making changes
func (r *Runner) DryRun() error {
	if r.Rewriter.Config.Skip {
		r.Logger.Println("Skipping dry run execution")
		return nil
	}

	// Load the environment
	err := r.Rewriter.LoadEnvironment()
	if err != nil {
		return fmt.Errorf("failed to load environment: %w", err)
	}

	// Get the build root
	buildRoot, err := r.Rewriter.GetBuildRoot()
	if err != nil {
		return fmt.Errorf("failed to get build root: %w", err)
	}

	r.Logger.Printf("Dry run - processing project at: %s", buildRoot)

	// Find source files
	sourceFiles, err := r.Rewriter.FindSourceFiles(buildRoot)
	if err != nil {
		return fmt.Errorf("failed to find source files: %w", err)
	}

	r.Logger.Printf("Found %d source files to process", len(sourceFiles))

	if len(sourceFiles) == 0 {
		r.Logger.Println("No source files found to process")
		return nil
	}

	// Process the files (but don't apply changes)
	results, err := r.Rewriter.ProcessFiles(sourceFiles)
	if err != nil {
		return fmt.Errorf("failed to process files: %w", err)
	}

	// Handle first exception if any
	if results.FirstException != nil {
		r.Logger.Printf("ERROR: The recipe produced an error: %v", results.FirstException)
		return results.FirstException
	}

	// Report what would be changed (but don't apply)
	if results.IsNotEmpty() {
		r.reportDryRunResults(results)
	} else {
		r.Logger.Println("No changes would be made")
	}

	return nil
}

// reportDryRunResults reports what would be changed in a dry run
func (r *Runner) reportDryRunResults(results *ResultsContainer) {
	r.Logger.Println("The following changes would be made:")

	if len(results.Generated) > 0 {
		r.Logger.Printf("Would generate %d new files:", len(results.Generated))
		for _, result := range results.Generated {
			if result.After != nil {
				r.Logger.Printf("  + %s", result.After.Path)
			}
		}
	}

	if len(results.Deleted) > 0 {
		r.Logger.Printf("Would delete %d files:", len(results.Deleted))
		for _, result := range results.Deleted {
			if result.Before != nil {
				r.Logger.Printf("  - %s", result.Before.Path)
			}
		}
	}

	if len(results.Moved) > 0 {
		r.Logger.Printf("Would move %d files:", len(results.Moved))
		for _, result := range results.Moved {
			if result.Before != nil && result.After != nil {
				r.Logger.Printf("  %s -> %s", result.Before.Path, result.After.Path)
			}
		}
	}

	if len(results.RefactoredInPlace) > 0 {
		r.Logger.Printf("Would modify %d files:", len(results.RefactoredInPlace))
		for _, result := range results.RefactoredInPlace {
			if result.Before != nil {
				r.Logger.Printf("  ~ %s", result.Before.Path)
			}
		}
	}

	r.Logger.Println("Run without --dry-run to apply these changes.")
}
