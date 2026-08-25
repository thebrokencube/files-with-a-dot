package lint

import (
	"go/ast"
	"regexp"
	"sort"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/conventions"
)

var goWorkUsePattern = regexp.MustCompile(`\./cmd/([a-zA-Z0-9_-]+)`)

//dendrik:block core-shell
//dendrik:kind component
//dendrik:layer bridge
//dendrik:status shipped
//dendrik:definition a tool's skill half wired to its importable Go core (go.work use, symlink_map, pkg core)
//dendrik:intent the skill-CLI pair is the reusable unit; wire it once, mechanically
//dendrik:conformance dendrik-import go-work-sync symlink-entries bundle-boundary core-in-pkg dispatch-router

// BridgeLint validates bridge layer conventions. Pure function — no I/O.
func BridgeLint(data *ToolData) []Result {
	var results []Result

	results = append(results, checkDendrikImport(data)...)
	results = append(results, checkBareReturnsAndOsExit(data)...)
	results = append(results, checkJSONFlagCoverage(data)...)
	results = append(results, checkGoWorkSync(data)...)
	results = append(results, checkSymlinkEntries(data)...)
	results = append(results, checkBundleBoundary(data)...)
	results = append(results, checkMakefileGofiles(data)...)
	results = append(results, checkNoJsonEncoder(data)...)
	results = append(results, checkNoRawJSONPassthrough(data)...)
	results = append(results, checkRunFunctionsHaveJSON(data)...)

	return results
}

func checkDendrikImport(data *ToolData) []Result {
	for _, gf := range data.GoFiles {
		if strings.Contains(string(gf.Content), "pkg/dendrik") {
			return nil
		}
	}
	return []Result{lintResult("dendrik-import", conventions.SeverityError,
		"no .go file imports pkg/dendrik",
		"", 0,
		"Import github.com/thebrokencube/files-with-a-dot/pkg/dendrik in at least one .go file.")}
}

func checkBareReturnsAndOsExit(data *ToolData) []Result {
	var results []Result

	for _, gf := range data.GoFiles {
		isCmdFile := strings.HasPrefix(gf.Path, "cmd_") && strings.HasSuffix(gf.Path, ".go")
		isMainFile := gf.Path == "main.go"

		if gf.AST == nil {
			continue
		}

		ast.Inspect(gf.AST, func(n ast.Node) bool {
			// Check bare integer returns in cmd_*.go
			if isCmdFile {
				ret, ok := n.(*ast.ReturnStmt)
				if ok && len(ret.Results) == 1 {
					lit, ok := ret.Results[0].(*ast.BasicLit)
					if ok && (lit.Value == "0" || lit.Value == "1" || lit.Value == "2" || lit.Value == "3") {
						results = append(results, lintResult("exit-constants", conventions.SeverityError,
							"bare return "+lit.Value+" — use dendrik.Exit* constants",
							gf.Path, 0,
							"Replace `return "+lit.Value+"` with the appropriate dendrik.Exit* constant."))
					}
				}
			}

			// Check os.Exit outside main.go
			if !isMainFile {
				call, ok := n.(*ast.CallExpr)
				if ok && isOsExitCall(call) {
					results = append(results, lintResult("exit-constants", conventions.SeverityError,
						"os.Exit() call outside main.go",
						gf.Path, 0,
						"Move os.Exit() calls to main.go only. Return exit codes from run*() functions."))
				}
			}

			return true
		})
	}

	return results
}

func checkJSONFlagCoverage(data *ToolData) []Result {
	// Per-tool granularity: if any file registers --json, at least one file uses WriteResult/WriteError
	hasJSONFlag := false
	hasWriteResult := false

	for _, gf := range data.GoFiles {
		content := string(gf.Content)
		if containsJSONFlagRegistration(content) {
			hasJSONFlag = true
		}
		if strings.Contains(content, "dendrik.WriteResult") || strings.Contains(content, "dendrik.WriteError") ||
			strings.Contains(content, ".Result(") || strings.Contains(content, ".MustResult(") {
			hasWriteResult = true
		}
	}

	if hasJSONFlag && !hasWriteResult {
		return []Result{lintResult("json-output", conventions.SeverityError,
			"--json flag registered but no dendrik.WriteResult/WriteError or Output.Result usage found",
			"", 0,
			"Add dendrik.WriteResult or Output.Result calls in commands that register a --json flag.")}
	}
	return nil
}

func checkGoWorkSync(data *ToolData) []Result {
	if data.GoWork == nil {
		return nil
	}

	var results []Result

	// Extract cmd/*/ entries from go.work
	goWorkCmds := map[string]bool{}
	for _, match := range goWorkUsePattern.FindAllStringSubmatch(string(data.GoWork), -1) {
		goWorkCmds[match[1]] = true
	}

	// Compare with actual cmd/*/ directories that have go.mod
	diskCmds := map[string]bool{}
	for _, name := range data.CmdDirs {
		diskCmds[name] = true
	}

	// In go.work but not on disk.
	goWorkNames := make([]string, 0, len(goWorkCmds))
	for name := range goWorkCmds {
		goWorkNames = append(goWorkNames, name)
	}
	sort.Strings(goWorkNames)
	for _, name := range goWorkNames {
		if !diskCmds[name] {
			results = append(results, lintResult("go-work-sync", conventions.SeverityError,
				"go.work references ./cmd/"+name+" but directory has no go.mod", "go.work", 0,
				"Remove `./cmd/"+name+"` from go.work or create go.mod in cmd/"+name+"/."))
		}
	}
	diskNames := make([]string, 0, len(diskCmds))
	for name := range diskCmds {
		diskNames = append(diskNames, name)
	}
	sort.Strings(diskNames)
	for _, name := range diskNames {
		if !goWorkCmds[name] {
			results = append(results, lintResult("go-work-sync", conventions.SeverityError,
				"cmd/"+name+"/ has go.mod but is not in go.work", "go.work", 0,
				"Add `./cmd/"+name+"` to the `use` block in go.work."))
		}
	}

	return results
}

func checkSymlinkEntries(data *ToolData) []Result {
	if data.SymlinkMap == nil {
		return nil // Opt-in: only runs when symlink_map.txt exists
	}
	expectedSource := "plugins/" + data.ToolName + "/skills/" + data.ToolName
	expectedDest := "$HOME/.claude/skills/" + data.ToolName
	found := false
	for _, line := range strings.Split(string(data.SymlinkMap), "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(parts) == 2 && parts[0] == expectedSource && parts[1] == expectedDest {
			found = true
			break
		}
	}
	if !found {
		return []Result{lintResult("symlink-entries", conventions.SeverityError,
			"symlink_map.txt missing exact skill entry for "+expectedSource,
			"symlink_map.txt", 0,
			"Add `"+expectedSource+":"+expectedDest+"`.")}
	}
	return nil
}
func checkBundleBoundary(data *ToolData) []Result {
	var results []Result
	if data.LegacyPluginPath {
		results = append(results, lintResult("bundle-boundary", conventions.SeverityError,
			"plugin delivery content remains under cmd/"+data.ToolName,
			"cmd/"+data.ToolName, 0,
			"Move skill, native manifest, and setup into plugins/"+data.ToolName+"."))
	}
	required := map[string]bool{
		".claude-plugin/plugin.json":            false,
		"bin/setup":                             false,
		"VERSION":                               false,
		"skills/" + data.ToolName + "/SKILL.md": false,
	}
	for _, path := range data.BundleFiles {
		if _, ok := required[path]; ok {
			required[path] = true
		}
		allowed := path == ".claude-plugin/plugin.json" || path == "bin/setup" || path == "VERSION" ||
			strings.HasPrefix(path, "skills/"+data.ToolName+"/")
		if !allowed {
			results = append(results, lintResult("bundle-boundary", conventions.SeverityError,
				"undeclared file in publishable bundle: "+path, path, 0,
				"Keep only the native manifest, canonical skill tree, bin/setup, and VERSION."))
		}
	}
	requiredPaths := make([]string, 0, len(required))
	for path := range required {
		requiredPaths = append(requiredPaths, path)
	}
	sort.Strings(requiredPaths)
	for _, path := range requiredPaths {
		if !required[path] {
			results = append(results, lintResult("bundle-boundary", conventions.SeverityError,
				"publishable bundle missing "+path, path, 0,
				"Add the required closed-bundle file."))
		}
	}
	return results
}

func checkMakefileGofiles(data *ToolData) []Result {
	if data.Makefile == nil {
		return nil
	}

	if !strings.Contains(string(data.Makefile), "../../pkg/dendrik") {
		return []Result{lintResult("makefile-gofiles", conventions.SeverityWarning,
			"Makefile GOFILES find path does not include ../../pkg/dendrik",
			"Makefile", 0,
			"Update Makefile GOFILES to: $(shell find . ../../pkg/dendrik -name '*.go')")}
	}
	return nil
}

func checkNoJsonEncoder(data *ToolData) []Result {
	var results []Result
	for _, gf := range data.GoFiles {
		if !strings.HasPrefix(gf.Path, "cmd_") {
			continue
		}
		if strings.Contains(string(gf.Content), "json.NewEncoder") {
			results = append(results, lintResult("no-json-encoder", conventions.SeverityError,
				"json.NewEncoder in "+gf.Path+" — bypasses ResultEnvelope contract",
				gf.Path, 0,
				"Replace json.NewEncoder with dendrik.Output.Result() or dendrik.WriteResult()."))
		}
	}
	return results
}

func checkNoRawJSONPassthrough(data *ToolData) []Result {
	var results []Result
	for _, gf := range data.GoFiles {
		if !strings.HasPrefix(gf.Path, "cmd_") {
			continue
		}
		content := string(gf.Content)
		if !containsJSONFlagRegistration(content) {
			continue
		}
		if hasRawJSONPassthrough(content) {
			results = append(results, lintResult("no-raw-json", conventions.SeverityWarning,
				"fmt.Print(string( in "+gf.Path+" with --json flag — raw JSON passthrough",
				gf.Path, 0,
				"Wrap in ResultEnvelope via dendrik.Output.Result(), or add //nolint:no-raw-json if intentional."))
		}
	}
	return results
}

// hasRawJSONPassthrough flags fmt.Print(string( patterns that bypass the
// ResultEnvelope. Already-enveloped lines (MustResult, .Result() and explicit
// //nolint:no-raw-json allowances are skipped to avoid false positives.
func hasRawJSONPassthrough(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		if !strings.Contains(line, "fmt.Print(string(") {
			continue
		}
		if strings.Contains(line, "MustResult") || strings.Contains(line, ".Result(") {
			continue
		}
		if strings.Contains(line, "//nolint:no-raw-json") {
			continue
		}
		return true
	}
	return false
}

func checkRunFunctionsHaveJSON(data *ToolData) []Result {
	var results []Result
	for _, gf := range data.GoFiles {
		if !strings.HasPrefix(gf.Path, "cmd_") {
			continue
		}
		content := string(gf.Content)
		// Check if file has run* functions
		if !strings.Contains(content, "func run") {
			continue
		}
		if !containsJSONFlagRegistration(content) {
			results = append(results, lintResult("run-has-json", conventions.SeverityWarning,
				gf.Path+" has run* function but no --json flag",
				gf.Path, 0,
				"Add a --json flag to the command's flag set, or acknowledge the gap is intentional."))
		}
	}
	return results
}

func containsJSONFlagRegistration(content string) bool {
	return strings.Contains(content, `"json"`) &&
		(strings.Contains(content, "BoolLong") || strings.Contains(content, "Bool("))
}
