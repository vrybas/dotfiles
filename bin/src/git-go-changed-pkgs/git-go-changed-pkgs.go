package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// PackageInfo represents information about a Go package
type PackageInfo struct {
	ImportPath string   // e.g., "github.com/Typeform/account-service/internal/errors"
	Dir        string   // Filesystem path
	Deps       []string // Import paths of dependencies (module-internal only)
	Imports    []string // All imports from go list
	Changed    bool     // Whether this package has changes
}

// DependencyGraph represents the dependency relationships between packages
type DependencyGraph struct {
	Packages map[string]*PackageInfo // All packages (changed + dependencies)
	AdjList  map[string][]string     // Reverse deps: pkg -> [packages that depend on pkg]
	InDegree map[string]int          // pkg -> count of its dependencies
}

// MergeGroup represents a group of packages that can be merged together
type MergeGroup struct {
	Level    int      // Merge order: 0 = no deps, 1 = depends on level 0, etc.
	Packages []string // Package import paths in this level
}

var (
	baseBranch = flag.String("base", "main", "Base branch to compare against")
	verbose    = flag.Bool("verbose", false, "Show detailed dependency information")
)

// Box drawing characters
const (
	boxHDouble    = "═"
	boxVDouble    = "║"
	boxCornerTL   = "╔"
	boxCornerTR   = "╗"
	boxCornerBL   = "╚"
	boxCornerBR   = "╝"
	boxHSingle    = "─"
	boxVSingle    = "│"
	boxCornerTLS  = "┌"
	boxCornerTRS  = "┐"
	boxCornerBLS  = "└"
	boxCornerBRS  = "┘"
	boxTeeRight   = "├"
	boxTeeEnd     = "└"
	arrowLeft     = "◀"
)

// Circled numbers for merge order markers
var circledNumbers = []string{"①", "②", "③", "④", "⑤", "⑥", "⑦", "⑧", "⑨", "⑩", "⑪", "⑫", "⑬", "⑭", "⑮"}

func main() {
	ctx := context.Background()
	flag.Parse()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// 1. Validate git repository
	if err := validateGitRepo(ctx); err != nil {
		return fmt.Errorf("git validation failed: %w", err)
	}

	// 2. Validate base branch exists
	if err := validateBaseBranch(ctx, *baseBranch); err != nil {
		return fmt.Errorf("base branch validation failed: %w", err)
	}

	// 3. Get current branch name
	currentBranch, err := getCurrentBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// 4. Get module prefix from go.mod
	modulePrefix, err := getModulePrefix(ctx)
	if err != nil {
		return fmt.Errorf("failed to get module prefix: %w", err)
	}

	// 5. Find changed packages
	changedPkgDirs, err := findChangedPackages(ctx, *baseBranch)
	if err != nil {
		return fmt.Errorf("failed to find changed packages: %w", err)
	}

	if len(changedPkgDirs) == 0 {
		fmt.Println("No Go files changed (committed, staged, or untracked)")
		return nil
	}

	// 6. Discover package information
	packages, err := discoverPackages(ctx, changedPkgDirs)
	if err != nil {
		return fmt.Errorf("failed to discover packages: %w", err)
	}

	// 7. Build dependency graph
	graph, err := buildDependencyGraph(packages, modulePrefix)
	if err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// 8. Perform topological sort
	mergeGroups, err := topologicalSort(graph, packages)
	if err != nil {
		return fmt.Errorf("topological sort failed: %w", err)
	}

	// 9. Output merge order
	outputMergeOrder(currentBranch, *baseBranch, modulePrefix, mergeGroups, graph, len(packages))

	return nil
}

// validateGitRepo checks if current directory is inside a git repository
func validateGitRepo(ctx context.Context) error {
	_, err := execCommand(ctx, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("not a git repository")
	}
	return nil
}

// validateBaseBranch checks if the base branch exists
func validateBaseBranch(ctx context.Context, branch string) error {
	_, err := execCommand(ctx, "git", "rev-parse", "--verify", branch)
	if err != nil {
		return fmt.Errorf("base branch %q does not exist", branch)
	}
	return nil
}

// getCurrentBranch returns the name of the current git branch
func getCurrentBranch(ctx context.Context) (string, error) {
	output, err := execCommand(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(output), nil
}

// getModulePrefix reads go.mod and returns the module path
func getModulePrefix(ctx context.Context) (string, error) {
	output, err := execCommand(ctx, "go", "list", "-m")
	if err != nil {
		return "", fmt.Errorf("failed to get module name: %w", err)
	}
	return strings.TrimSpace(output), nil
}

// addGoFilesToPkgDirs parses git command output and adds .go file directories to pkgDirs
func addGoFilesToPkgDirs(output string, pkgDirs map[string]bool) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasSuffix(line, ".go") {
			continue
		}
		dir := filepath.Dir(line)
		if dir == "." {
			dir = "./"
		} else {
			dir = "./" + dir
		}
		pkgDirs[dir] = true
	}
}

// findChangedPackages finds all packages that have .go files changed (committed, staged, or untracked)
func findChangedPackages(ctx context.Context, baseBranch string) ([]string, error) {
	pkgDirs := make(map[string]bool)

	// 1. Get committed changes (branch vs base)
	output, err := execCommand(ctx, "git", "log", "--first-parent", "--no-merges", "--format=", "--name-only", baseBranch+"..HEAD")
	if err != nil {
		return nil, fmt.Errorf("git log committed failed: %w", err)
	}
	addGoFilesToPkgDirs(output, pkgDirs)

	// 2. Get staged changes (not yet committed)
	output, err = execCommand(ctx, "git", "diff", "--name-only", "--cached")
	if err != nil {
		return nil, fmt.Errorf("git diff staged failed: %w", err)
	}
	addGoFilesToPkgDirs(output, pkgDirs)

	// 3. Get untracked files
	output, err = execCommand(ctx, "git", "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, fmt.Errorf("git ls-files failed: %w", err)
	}
	addGoFilesToPkgDirs(output, pkgDirs)

	// Convert map to sorted slice
	result := make([]string, 0, len(pkgDirs))
	for dir := range pkgDirs {
		result = append(result, dir)
	}
	sort.Strings(result)
	return result, nil
}

// discoverPackages analyzes the given package directories and returns package information
func discoverPackages(ctx context.Context, pkgDirs []string) (map[string]*PackageInfo, error) {
	packages := make(map[string]*PackageInfo)

	for _, dir := range pkgDirs {
		// Run go list -json for this directory
		output, err := execCommand(ctx, "go", "list", "-json", dir)
		if err != nil {
			return nil, fmt.Errorf("go list failed for %s: %w", dir, err)
		}

		// Parse JSON output
		pkgInfos, err := parseGoListJSON(output)
		if err != nil {
			return nil, fmt.Errorf("failed to parse go list output for %s: %w", dir, err)
		}

		if len(pkgInfos) == 0 {
			continue
		}

		// Take the first (and should be only) package
		pkg := pkgInfos[0]
		pkg.Changed = true
		packages[pkg.ImportPath] = pkg
	}

	return packages, nil
}

// buildDependencyGraph builds a dependency graph for the changed packages
func buildDependencyGraph(changedPackages map[string]*PackageInfo, modulePrefix string) (*DependencyGraph, error) {
	graph := &DependencyGraph{
		Packages: make(map[string]*PackageInfo),
		AdjList:  make(map[string][]string),
		InDegree: make(map[string]int),
	}

	// Initialize graph with changed packages
	for importPath, pkg := range changedPackages {
		graph.Packages[importPath] = pkg
		graph.AdjList[importPath] = []string{}
		graph.InDegree[importPath] = 0
	}

	// For each changed package, analyze its dependencies
	for importPath, pkg := range changedPackages {
		// Filter to module-internal dependencies only
		internalDeps := filterModulePackages(pkg.Imports, modulePrefix, changedPackages)
		pkg.Deps = internalDeps

		// Build reverse dependency graph (adjacency list)
		// If A imports B, then B's adjacency list includes A
		for _, depPath := range internalDeps {
			if graph.AdjList[depPath] == nil {
				graph.AdjList[depPath] = []string{}
			}
			graph.AdjList[depPath] = append(graph.AdjList[depPath], importPath)
			graph.InDegree[importPath]++
		}
	}

	return graph, nil
}

// filterModulePackages filters imports to only include packages from the same module that are in changedPackages
func filterModulePackages(imports []string, modulePrefix string, changedPackages map[string]*PackageInfo) []string {
	var result []string

	for _, imp := range imports {
		// Only include packages from the same module
		if !strings.HasPrefix(imp, modulePrefix) {
			continue
		}

		// Only include packages that are in the changed set
		if _, exists := changedPackages[imp]; !exists {
			continue
		}

		result = append(result, imp)
	}

	return result
}

// topologicalSort performs Kahn's algorithm to compute merge order
func topologicalSort(graph *DependencyGraph, changedPackages map[string]*PackageInfo) ([]MergeGroup, error) {
	// Create a working copy of in-degrees
	inDegree := make(map[string]int)
	for pkg, degree := range graph.InDegree {
		inDegree[pkg] = degree
	}

	// Initialize queue with packages having in-degree 0
	var queue []string
	for pkg := range changedPackages {
		if inDegree[pkg] == 0 {
			queue = append(queue, pkg)
		}
	}

	// Sort queue for consistent output
	sort.Strings(queue)

	var mergeGroups []MergeGroup
	level := 0

	for len(queue) > 0 {
		// Current level contains all packages in the queue
		currentLevel := MergeGroup{
			Level:    level,
			Packages: append([]string{}, queue...),
		}
		mergeGroups = append(mergeGroups, currentLevel)

		// Process all packages in current queue
		var nextQueue []string
		for _, pkg := range queue {
			// For each package that depends on this one
			for _, dependent := range graph.AdjList[pkg] {
				inDegree[dependent]--
				if inDegree[dependent] == 0 {
					nextQueue = append(nextQueue, dependent)
				}
			}
		}

		// Sort next queue for consistent output
		sort.Strings(nextQueue)
		queue = nextQueue
		level++
	}

	// Check for cycles (packages with in-degree > 0 after sort)
	var cyclePackages []string
	for pkg := range changedPackages {
		if inDegree[pkg] > 0 {
			cyclePackages = append(cyclePackages, pkg)
		}
	}

	if len(cyclePackages) > 0 {
		sort.Strings(cyclePackages)
		return nil, fmt.Errorf("circular dependencies detected among packages: %v", cyclePackages)
	}

	return mergeGroups, nil
}

// repeatString repeats a string n times
func repeatString(s string, n int) string {
	if n <= 0 {
		return ""
	}
	var builder strings.Builder
	for i := 0; i < n; i++ {
		builder.WriteString(s)
	}
	return builder.String()
}

// getCircledNumber returns a circled number marker for the given index (0-based)
func getCircledNumber(idx int) string {
	if idx < len(circledNumbers) {
		return circledNumbers[idx]
	}
	return fmt.Sprintf("[%d]", idx+1)
}

// drawHeader draws a double-line box header
func drawHeader(title string, width int) {
	// Ensure minimum width
	if width < len(title)+4 {
		width = len(title) + 4
	}
	innerWidth := width - 2

	fmt.Printf("%s%s%s\n", boxCornerTL, repeatString(boxHDouble, innerWidth), boxCornerTR)
	// Pad title to center
	padding := (innerWidth - len(title)) / 2
	rightPadding := innerWidth - len(title) - padding
//	fmt.Printf("%s%s%s%s%s\n", boxVDouble, repeatString(" ", padding), title, repeatString(" ", rightPadding), boxVDouble)
	fmt.Printf("%s%s%s%s\n", boxVDouble, repeatString(" ", padding), title, repeatString(" ", rightPadding))
	fmt.Printf("%s%s%s\n", boxCornerBL, repeatString(boxHDouble, innerWidth), boxCornerBR)
}

// drawSectionHeader draws a single-line section header
func drawSectionHeader(title string) {
	width := 65
	innerWidth := width - 2

	fmt.Printf("%s%s%s\n", boxCornerTLS, repeatString(boxHSingle, innerWidth), boxCornerTRS)
	fmt.Printf("%s %s%s%s\n", boxVSingle, title, repeatString(" ", innerWidth-len(title)-1), boxVSingle)
	fmt.Printf("%s%s%s\n", boxCornerBLS, repeatString(boxHSingle, innerWidth), boxCornerBRS)
}

// buildPackageNumberMap creates a map from package path to its merge order number
func buildPackageNumberMap(groups []MergeGroup) map[string]int {
	pkgNum := make(map[string]int)
	num := 0
	for _, group := range groups {
		for _, pkg := range group.Packages {
			pkgNum[pkg] = num
			num++
		}
	}
	return pkgNum
}

// renderDependencyTree renders the bottom-up dependency tree
func renderDependencyTree(modulePrefix string, groups []MergeGroup, graph *DependencyGraph) {
	drawSectionHeader("DEPENDENCY TREE (merge from top to bottom)")
	fmt.Println()

	pkgNum := buildPackageNumberMap(groups)

	// Track which packages have already been rendered as children
	rendered := make(map[string]bool)

	// Start with level 0 packages (no dependencies) as roots
	if len(groups) == 0 {
		return
	}

	// Render each root and its dependents
	for _, rootPkg := range groups[0].Packages {
		renderTreeNode(rootPkg, modulePrefix, graph, pkgNum, rendered, "", true, true)
	}

	fmt.Println()
	fmt.Println("  Legend: Parent ← Child means \"Child depends on Parent\"")
	fmt.Println()
}

// renderTreeNode recursively renders a node and its dependents
func renderTreeNode(pkg, modulePrefix string, graph *DependencyGraph, pkgNum map[string]int, rendered map[string]bool, prefix string, isLast bool, isRoot bool) {
	shortName := strings.TrimPrefix(pkg, modulePrefix+"/")
	marker := getCircledNumber(pkgNum[pkg])

	// Build the connection prefix
	var connector string
	if isRoot {
		connector = "  "
	} else if isLast {
		connector = boxTeeEnd + boxHSingle + boxHSingle + " "
	} else {
		connector = boxTeeRight + boxHSingle + boxHSingle + " "
	}

	// Determine annotation
	var annotation string
	pkgInfo := graph.Packages[pkg]
	if len(pkgInfo.Deps) == 0 {
		annotation = fmt.Sprintf(" %s%s no dependencies", arrowLeft, boxHSingle)
	} else {
		depMarkers := make([]string, 0, len(pkgInfo.Deps))
		for _, dep := range pkgInfo.Deps {
			depMarkers = append(depMarkers, getCircledNumber(pkgNum[dep]))
		}
		sort.Strings(depMarkers)
		annotation = fmt.Sprintf(" %s%s depends on %s", arrowLeft, boxHSingle, strings.Join(depMarkers, ""))
	}

	fmt.Printf("%s%s%s %s%s\n", prefix, connector, marker, shortName, annotation)

	// Mark as rendered
	rendered[pkg] = true

	// Get dependents (packages that depend on this one)
	dependents := append([]string{}, graph.AdjList[pkg]...)
	sort.Strings(dependents)

	// Filter to only unrendered dependents to avoid duplication
	var unrenderedDeps []string
	for _, dep := range dependents {
		if !rendered[dep] {
			unrenderedDeps = append(unrenderedDeps, dep)
		}
	}

	// Build new prefix for children
	var childPrefix string
	if isRoot {
		childPrefix = prefix + "     "
	} else if isLast {
		childPrefix = prefix + "    "
	} else {
		childPrefix = prefix + boxVSingle + "   "
	}

	// Render children
	for i, dep := range unrenderedDeps {
		isLastChild := (i == len(unrenderedDeps)-1)
		renderTreeNode(dep, modulePrefix, graph, pkgNum, rendered, childPrefix, isLastChild, false)
	}
}

// renderMergeOrder renders the numbered merge order list
func renderMergeOrder(modulePrefix string, groups []MergeGroup, graph *DependencyGraph) {
	drawSectionHeader("MERGE ORDER")
	fmt.Println()

	pkgNum := buildPackageNumberMap(groups)
	num := 0

	for _, group := range groups {
		for _, pkg := range group.Packages {
			shortName := strings.TrimPrefix(pkg, modulePrefix+"/")
			marker := getCircledNumber(num)
			pkgInfo := graph.Packages[pkg]

			var depInfo string
			if len(pkgInfo.Deps) == 0 {
				depInfo = "merge first (no deps)"
			} else {
				depMarkers := make([]string, 0, len(pkgInfo.Deps))
				for _, dep := range pkgInfo.Deps {
					depMarkers = append(depMarkers, getCircledNumber(pkgNum[dep]))
				}
				sort.Strings(depMarkers)
				depInfo = fmt.Sprintf("after %s", strings.Join(depMarkers, ""))
			}

			// Calculate padding for alignment
			maxNameLen := 35
			padLen := maxNameLen - len(shortName)
			if padLen < 2 {
				padLen = 2
			}

			fmt.Printf("  %s %s%s%s\n", marker, shortName, repeatString(" ", padLen), depInfo)
			num++
		}
	}
	fmt.Println()
}

// outputMergeOrder prints the merge order in a human-readable format with ASCII art
func outputMergeOrder(currentBranch, baseBranch, modulePrefix string, groups []MergeGroup, graph *DependencyGraph, totalPackages int) {
	// Header banner
	title := fmt.Sprintf("Package Dependencies: %s → %s", currentBranch, baseBranch)
	subtitle := fmt.Sprintf("Changed: %d packages · %d merge levels", totalPackages, len(groups))
	headerWidth := 68

	fmt.Println()
	drawHeader(title, headerWidth)

	// Subtitle line
	subtitlePad := (headerWidth - len(subtitle) - 2) / 2
	fmt.Printf("%s%s\n\n", repeatString(" ", subtitlePad), subtitle)

	// Dependency tree visualization
	// renderDependencyTree(modulePrefix, groups, graph)

	// Numbered merge order
	renderMergeOrder(modulePrefix, groups, graph)

	// Summary
	fmt.Printf("  To split this PR, create %d separate PRs in the order above.\n", len(groups))
	fmt.Println()

	if *verbose {
		drawSectionHeader("VERBOSE DETAILS")
		fmt.Println()
		fmt.Printf("  Module prefix: %s\n", modulePrefix)
		fmt.Printf("  Total packages in graph: %d\n", len(graph.Packages))
		fmt.Printf("  Total dependency edges: %d\n", countEdges(graph))
		fmt.Println()
	}
}

// countEdges counts total number of dependency edges in the graph
func countEdges(graph *DependencyGraph) int {
	count := 0
	for _, adjacents := range graph.AdjList {
		count += len(adjacents)
	}
	return count
}

// parseGoListJSON parses the JSON output from go list -json
func parseGoListJSON(output string) ([]*PackageInfo, error) {
	var result []*PackageInfo

	decoder := json.NewDecoder(strings.NewReader(output))
	for decoder.More() {
		var pkg struct {
			ImportPath string   `json:"ImportPath"`
			Dir        string   `json:"Dir"`
			Imports    []string `json:"Imports"`
		}

		if err := decoder.Decode(&pkg); err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w", err)
		}

		result = append(result, &PackageInfo{
			ImportPath: pkg.ImportPath,
			Dir:        pkg.Dir,
			Imports:    pkg.Imports,
			Deps:       []string{},
		})
	}

	return result, nil
}

// execCommand executes a command and returns its stdout
func execCommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command %s %v failed: %w\nOutput: %s", name, args, err, string(output))
	}
	return string(output), nil
}
