//go:build mage

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Color functions using ANSI escape codes
func colorize(color, text string) string {
	return fmt.Sprintf("\033[%sm%s\033[0m", color, text)
}

func Header(text string) string    { return colorize("1;36", text) } // Bold Cyan
func InfoStyle(text string) string { return colorize("34", text) }   // Blue
func Success(text string) string   { return colorize("32", text) }   // Green
func Error(text string) string     { return colorize("31", text) }   // Red
func Warning(text string) string   { return colorize("33", text) }   // Yellow
func Dim(text string) string       { return colorize("2", text) }    // Dim
func Code(text string) string      { return colorize("36", text) }   // Cyan

// Example formats a command and description
func Example(cmd, desc string) string {
	return fmt.Sprintf("  %s - %s", colorize("36", cmd), desc)
}

// Info displays the available mage commands and their descriptions
func Info() {
	fmt.Println(Header("Mage build script for vulpo"))
	fmt.Println()
	fmt.Println(InfoStyle("Available commands:"))
	fmt.Println()
	fmt.Println(InfoStyle("🧪 Quality Commands:"))
	fmt.Println(Example("ci", "Run full CI pipeline (format, test, lint)"))
	fmt.Println(Example("test", "Run all tests"))
	fmt.Println(Example("lint", "Run golangci-lint on project code"))
	fmt.Println(Example("lintall", "Run golangci-lint on project code and magefiles"))
	fmt.Println(Example("format", "Format Go code using gofmt"))
	fmt.Println()
	fmt.Println(InfoStyle("📋 Version & Release:"))
	fmt.Println(Example("version", "Display current version from VERSION file"))
	fmt.Println(Example("release", "Create and push annotated release tag"))
	fmt.Println(Example("publish", "Run checks, build, and publish new release"))
	fmt.Println()
	fmt.Println(InfoStyle("🔍 Git Commands:"))
	fmt.Println(Example("git:committed", "Check if git repository has no uncommitted changes"))
	fmt.Println(Example("git:pushed", "Check if all commits are pushed to remote"))
	fmt.Println()
	fmt.Printf("%s %s\n", InfoStyle("Usage:"), "mage <command>")
	fmt.Println(Dim("Examples:"))
	fmt.Printf("%s %s\n", Dim("  Run CI pipeline:"), Code("mage ci"))
	fmt.Printf("%s %s\n", Dim("  Check version:"), Code("mage version"))
	fmt.Printf("%s %s\n", Dim("  Create release:"), Code("mage release"))
	fmt.Printf("%s %s\n", Dim("  Full publish:"), Code("mage publish"))
	fmt.Printf("%s %s\n", Dim("Tip:"), "Run 'mage -l' to list all available commands")
	fmt.Println()
	fmt.Println(Success("Ready to go!"))
}

// CI runs the full CI pipeline: format, test, lint
func CI() error {
	fmt.Println(Header("🚀 Running CI pipeline..."))
	fmt.Println()

	// Format code first
	if err := Format(); err != nil {
		return err
	}
	fmt.Println()

	// Run tests
	if err := Test(); err != nil {
		return err
	}
	fmt.Println()

	// Run linting last
	if err := Lint(); err != nil {
		return err
	}
	fmt.Println()

	fmt.Println(Success("🎉 CI pipeline completed successfully!"))
	return nil
}

// Test runs all tests in the project
func Test() error {
	fmt.Println(InfoStyle("Running tests..."))

	if err := sh.RunV("go", "test", "./...", "-count=1"); err != nil {
		return fmt.Errorf("%s tests failed: %v", Error("Error:"), err)
	}

	fmt.Println(Success("✓ All tests passed"))
	return nil
}

// Lint runs golangci-lint on the project (excludes magefiles)
func Lint() error {
	fmt.Println(InfoStyle("Running golangci-lint..."))

	if err := sh.RunV("golangci-lint", "run"); err != nil {
		return fmt.Errorf("%s linting failed: %v", Error("Error:"), err)
	}

	fmt.Println(Success("✓ Linting completed successfully"))
	return nil
}

// LintAll runs golangci-lint on the project including magefiles
func LintAll() error {
	fmt.Println(InfoStyle("Running golangci-lint (including magefiles)..."))

	if err := sh.RunV("golangci-lint", "run", "--build-tags=mage"); err != nil {
		return fmt.Errorf("%s linting failed: %v", Error("Error:"), err)
	}

	fmt.Println(Success("✓ Linting completed successfully"))
	return nil
}

// Format runs gofmt on all Go files in the project
func Format() error {
	fmt.Println(InfoStyle("Formatting Go code with gofmt..."))

	if err := sh.RunV("gofmt", "-s", "-w", "."); err != nil {
		return fmt.Errorf("%s formatting failed: %v", Error("Error:"), err)
	}

	fmt.Println(Success("✓ Code formatting completed successfully"))
	return nil
}

// GetVersion reads the version from the VERSION file
func GetVersion() (string, error) {
	data, err := os.ReadFile("VERSION")
	if err != nil {
		return "", fmt.Errorf("failed to read VERSION file: %w", err)
	}

	version := strings.TrimSpace(string(data))
	if version == "" {
		return "", fmt.Errorf("VERSION file is empty")
	}

	// Validate version format (semantic versioning)
	if matched, _ := regexp.MatchString(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`, version); !matched {
		return "", fmt.Errorf("invalid version format: %s (expected format: x.y.z or vx.y.z)", version)
	}

	// Ensure version starts with 'v'
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	return version, nil
}

// Version displays the current version
func Version() error {
	version, err := GetVersion()
	if err != nil {
		return err
	}

	fmt.Printf("%s Current version: %s\n", InfoStyle("📋"), Success(version))
	return nil
}

// CheckGitStatus ensures the git repository is in a clean state
func CheckGitStatus() error {
	// Check for uncommitted changes
	output, err := sh.Output("git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if strings.TrimSpace(output) != "" {
		return fmt.Errorf("%s repository has uncommitted changes. Commit or stash changes before publishing", Error("Error:"))
	}

	// Check if we're on main/master branch
	currentBranch, err := sh.Output("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	currentBranch = strings.TrimSpace(currentBranch)
	if currentBranch != "main" && currentBranch != "master" {
		fmt.Printf("%s Warning: Publishing from branch '%s' (not main/master)\n", Warning("⚠️"), currentBranch)
		fmt.Print("Continue? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			return fmt.Errorf("publish cancelled")
		}
	}

	return nil
}

// CheckVersionBump ensures the version has been bumped since the last tag
func CheckVersionBump(version string) error {
	// Get the latest tag
	latestTag, err := sh.Output("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		// No existing tags, this is the first release
		fmt.Println(InfoStyle("📋 No existing tags found - this will be the first release"))
		return nil
	}

	latestTag = strings.TrimSpace(latestTag)

	// Check if version already exists as a tag
	if latestTag == version {
		return fmt.Errorf("%s version %s already exists as a git tag. Please bump the version in the VERSION file", Error("Error:"), version)
	}

	fmt.Printf("%s Version bump detected: %s → %s\n", InfoStyle("📋"), Dim(latestTag), Success(version))
	return nil
}

// Release creates and pushes a new release tag
func Release() error {
	fmt.Println(Header("🚀 Creating release..."))
	fmt.Println()

	// Get version
	version, err := GetVersion()
	if err != nil {
		return err
	}

	// Check git status
	if err := CheckGitStatus(); err != nil {
		return err
	}

	// Check version bump
	if err := CheckVersionBump(version); err != nil {
		return err
	}

	return createAndPushTag(version)
}

// createAndPushTag creates and pushes a Git tag with annotation
func createAndPushTag(version string) error {
	// Create annotated tag
	fmt.Printf("%s Creating annotated tag %s...\n", InfoStyle("🏷️"), Success(version))
	if err := sh.Run("git", "tag", "-a", version, "-m", fmt.Sprintf("Release %s", version)); err != nil {
		return fmt.Errorf("%s failed to create tag: %w", Error("Error:"), err)
	}

	fmt.Printf("%s Pushing tag %s...\n", InfoStyle("📤"), Success(version))
	if err := sh.Run("git", "push", "origin", version); err != nil {
		return fmt.Errorf("%s failed to push tag: %w", Error("Error:"), err)
	}

	fmt.Printf("%s Release %s created and pushed successfully!\n", Success("✅"), Success(version))
	return nil
}

// Publish runs all checks, creates a release, and pushes to remote
func Publish() error {
	fmt.Println(Header("🚀 Starting publish process..."))
	fmt.Println()

	// Get version first to display it
	version, err := GetVersion()
	if err != nil {
		return err
	}

	fmt.Printf("%s Publishing version: %s\n", InfoStyle("📋"), Success(version))
	fmt.Println()

	// Run all pre-publish checks
	fmt.Println(InfoStyle("📋 Running pre-publish checks..."))
	mg.SerialDeps(CheckGitStatus, CI)

	// Check version bump
	if err := CheckVersionBump(version); err != nil {
		return err
	}

	// Push current changes
	fmt.Println(InfoStyle("📤 Pushing current branch..."))
	if err := sh.Run("git", "push"); err != nil {
		return fmt.Errorf("%s failed to push current branch: %w", Error("Error:"), err)
	}

	// Create and push tag
	fmt.Printf("%s Creating annotated tag %s...\n", InfoStyle("🏷️"), Success(version))
	if err := sh.Run("git", "tag", "-a", version, "-m", fmt.Sprintf("Release %s", version)); err != nil {
		return fmt.Errorf("%s failed to create tag: %w", Error("Error:"), err)
	}

	fmt.Printf("%s Pushing tag %s...\n", InfoStyle("📤"), Success(version))
	if err := sh.Run("git", "push", "origin", version); err != nil {
		return fmt.Errorf("%s failed to push tag: %w", Error("Error:"), err)
	}

	fmt.Printf("%s Release %s created and pushed successfully!\n", Success("✅"), Success(version))
	fmt.Println()
	fmt.Printf("%s Successfully published %s!\n", Success("🎉"), Success(version))
	fmt.Printf("%s View release: https://github.com/mkfoss/vulpo/releases/tag/%s\n", InfoStyle("🔗"), version)
	fmt.Println()

	return nil
}

// Git namespace for git-related commands
type Git mg.Namespace

// Committed checks if the git repository has no uncommitted changes
func (Git) Committed() error {
	if !isGitClean() {
		return fmt.Errorf("%s repository is not clean", Error("Error:"))
	}
	return nil
}

// Pushed checks if all commits have been pushed to the remote repository
func (Git) Pushed() error {
	if !isGitPushed() {
		return fmt.Errorf("%s there are unpushed commits", Error("Error:"))
	}
	return nil
}

// Helper functions for git status checking
func isGitClean() bool {
	output, err := sh.Output("git", "status", "--porcelain")
	if err != nil {
		return false
	}
	return strings.TrimSpace(output) == ""
}

func isGitPushed() bool {
	// Check if there are unpushed commits
	output, err := sh.Output("git", "log", "--oneline", "@{u}..")
	if err != nil {
		// If we can't get the upstream, assume we need to push
		return false
	}
	return strings.TrimSpace(output) == ""
}

// Default target to run when no target is specified
var Default = Info
