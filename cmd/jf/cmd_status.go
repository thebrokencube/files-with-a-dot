package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/engine"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	gh "github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/github"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

// prFetcher is the function used to fetch PRs. Tests override this.
var prFetcher gh.Fetcher = gh.DefaultFetcher

func runStatus(dir string, jsonOut bool) int {
	f, roots, code := loadForestOrFail(dir, jsonOut)
	if code != 0 {
		return code
	}

	all := forest.Flatten(roots)

	state, err := forest.LoadState(f.Dir)
	if err != nil {
		state = &forest.State{Nodes: make(map[string]forest.NodeState)}
	}

	var pushPullCount, pullOnlyCount, pushOnlyCount, tbdTotal int
	var pushStale, pullStale int
	var mutableCount, readOnlyCount, emptyCount int

	type readOnlyDetail struct {
		Key       string
		Issue     string
		Effective string
	}
	var readOnlyDetails []readOnlyDetail

	for _, n := range all {
		if forest.IsTBD(n.Key) {
			tbdTotal++
			continue
		}

		// Explicit pull-only override
		if n.Sync == "pull" {
			pullOnlyCount++
			if state.IsPullStale(n.Key) {
				pullStale++
			}
			continue
		}

		// For push and both: check file existence and content
		filePath := filepath.Join(f.Dir, n.File)
		info, statErr := os.Stat(filePath)

		// Explicit push-only override
		if n.Sync == "push" {
			pushOnlyCount++
			if statErr != nil {
				pushStale++
				emptyCount++
				continue
			}
			if state.IsStale(n.Key, info.ModTime()) {
				pushStale++
			}
			// Mutability check — engine blocks read-only push nodes
			source, readErr := os.ReadFile(filePath)
			if readErr != nil {
				emptyCount++
				continue
			}
			stripped := pipeline.StripFrontmatter(source)
			if !engine.IsSubstantiveLocal(stripped) {
				emptyCount++
				continue
			}
			issues := pipeline.Lint(stripped, n.File)
			if len(issues) > 0 {
				readOnlyCount++
				readOnlyDetails = append(readOnlyDetails, readOnlyDetail{
					Key: n.Key, Issue: issues[0].Message, Effective: "blocked (read-only, sync:push)",
				})
			}
			continue
		}

		// Default case: sync == "both" — derive effective direction from mutability
		if statErr != nil {
			emptyCount++
			continue
		}

		source, readErr := os.ReadFile(filePath)
		if readErr != nil {
			emptyCount++
			continue
		}
		stripped := pipeline.StripFrontmatter(source)
		if !engine.IsSubstantiveLocal(stripped) {
			emptyCount++
			continue
		}

		// Mutability check (lint only — no Node.js roundtrip in status)
		issues := pipeline.Lint(stripped, n.File)
		if len(issues) > 0 {
			pullOnlyCount++
			readOnlyCount++
			if state.IsPullStale(n.Key) {
				pullStale++
			}
			readOnlyDetails = append(readOnlyDetails, readOnlyDetail{
				Key: n.Key, Issue: issues[0].Message, Effective: "pull-only (demoted)",
			})
		} else if cached, ok := state.MutabilityCache(n.Key, pipeline.ComputeLocalHash(stripped)); ok {
			if cached {
				pushPullCount++
				mutableCount++
				if state.IsStale(n.Key, info.ModTime()) {
					pushStale++
				}
			} else {
				pullOnlyCount++
				readOnlyCount++
				if state.IsPullStale(n.Key) {
					pullStale++
				}
				readOnlyDetails = append(readOnlyDetails, readOnlyDetail{
					Key: n.Key, Issue: "roundtrip check failed", Effective: "pull-only (demoted)",
				})
			}
		} else {
			pushPullCount++
			mutableCount++ // lint passes, no cache — assume mutable
			if state.IsStale(n.Key, info.ModTime()) {
				pushStale++
			}
		}
	}

	// PR enrichment: fetch PRs if repos are configured
	var prBadges []output.PRBadge
	repos := f.Defaults.Repos
	if len(repos) > 0 && prFetcher != nil {
		// Collect non-TBD keys for the search query
		var keys []string
		for _, n := range all {
			if !forest.IsTBD(n.Key) {
				keys = append(keys, n.Key)
			}
		}
		allPRs, fetchErr := prFetcher(repos, keys)
		if fetchErr == nil && allPRs != nil {
			for _, n := range all {
				if forest.IsTBD(n.Key) {
					continue
				}

				// Check frontmatter for manual prs: override
				filePath := filepath.Join(f.Dir, n.File)
				content, readErr := os.ReadFile(filePath)
				var manualPRs []int
				if readErr == nil {
					fm, fmErr := forest.ParseFrontmatter(content)
					if fmErr == nil && fm != nil && len(fm.PRs) > 0 {
						manualPRs = fm.PRs
					}
				}

				if len(manualPRs) > 0 {
					// Manual override: find these PR numbers in the fetched list
					for _, num := range manualPRs {
						found := false
						for _, pr := range allPRs {
							if pr.Number == num {
								prBadges = append(prBadges, output.PRBadge{
									Key:    n.Key,
									Number: pr.Number,
									State:  gh.DeriveState(pr),
									CI:     gh.DeriveCIStatus(pr.StatusCheckRollup),
									Branch: pr.HeadRefName,
								})
								found = true
								break
							}
						}
						if !found {
							// PR not in fetched list — show with unknown state
							prBadges = append(prBadges, output.PRBadge{
								Key:    n.Key,
								Number: num,
								State:  "unknown",
								CI:     "none",
							})
						}
					}
				} else {
					// Derive by matching
					matched := gh.MatchPRs(allPRs, n.Key)
					for _, pr := range matched {
						prBadges = append(prBadges, output.PRBadge{
							Key:    n.Key,
							Number: pr.Number,
							State:  gh.DeriveState(pr),
							CI:     gh.DeriveCIStatus(pr.StatusCheckRollup),
							Branch: pr.HeadRefName,
						})
					}
				}
			}
		}
	}

	// JSON output: PushTotal includes empty/blocked "both" nodes for backward compat
	// (old code counted all non-pull nodes as push)
	if jsonOut {
		dendrik.WriteResult(os.Stdout, output.StatusResult{
			Forest:    filepath.Dir(f.Dir),
			Total:     len(all),
			TBD:       tbdTotal,
			PushTotal: pushPullCount + pushOnlyCount + emptyCount,
			PushStale: pushStale,
			PullTotal: pullOnlyCount,
			PullStale: pullStale,
			Mutable:   mutableCount,
			ReadOnly:  readOnlyCount,
			Empty:     emptyCount,
			Repos:     repos,
			PRs:       prBadges,
		})
		return dendrik.ExitOK
	}

	fmt.Printf("Forest: %s\n", filepath.Dir(f.Dir))
	fmt.Printf("Nodes:  %d total", len(all))
	if tbdTotal > 0 {
		fmt.Printf(" (%d TBD)", tbdTotal)
	}
	fmt.Println()

	fmt.Println()
	fmt.Println("Effective direction:")
	if pushPullCount > 0 {
		fmt.Printf("  %d push+pull (mutable)\n", pushPullCount)
	}
	if pullOnlyCount > 0 {
		fmt.Printf("  %d pull-only", pullOnlyCount)
		if readOnlyCount > 0 {
			fmt.Printf(" (%d read-only demoted)", readOnlyCount)
		}
		fmt.Println()
	}
	if pushOnlyCount > 0 {
		fmt.Printf("  %d push-only (explicit override)\n", pushOnlyCount)
	}
	if emptyCount > 0 {
		fmt.Printf("  %d empty\n", emptyCount)
	}

	if pushStale > 0 || pullStale > 0 {
		fmt.Println()
		fmt.Println("Stale:")
		if pushStale > 0 {
			fmt.Printf("  %d push-eligible nodes changed locally\n", pushStale)
		}
		if pullStale > 0 {
			fmt.Printf("  %d pull-eligible nodes changed remotely\n", pullStale)
		}
	}

	if len(readOnlyDetails) > 0 {
		fmt.Println()
		fmt.Println("Read-only nodes:")
		for _, d := range readOnlyDetails {
			fmt.Printf("  %-10s %s — %s\n", d.Key, d.Issue, d.Effective)
		}
	}

	if len(repos) > 0 && len(prBadges) > 0 {
		fmt.Println()
		fmt.Printf("PRs (repos: %s):\n", strings.Join(repos, ", "))
		// Group by key, show each PR
		seen := make(map[string]bool)
		for _, n := range all {
			if forest.IsTBD(n.Key) || seen[n.Key] {
				continue
			}
			seen[n.Key] = true
			var nodePRs []output.PRBadge
			for _, b := range prBadges {
				if b.Key == n.Key {
					nodePRs = append(nodePRs, b)
				}
			}
			if len(nodePRs) == 0 {
				continue
			}
			for _, b := range nodePRs {
				ciStr := ""
				if b.CI != "none" && b.CI != "" {
					ciStr = fmt.Sprintf(" CI:%s", b.CI)
				}
				fmt.Printf("  %-12s PR #%d [%s]%s\n", b.Key, b.Number, b.State, ciStr)
			}
		}
	}

	return dendrik.ExitOK
}
