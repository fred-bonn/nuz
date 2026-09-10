package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/fred-bonn/nuz/internal/engine"
	"github.com/fred-bonn/nuz/internal/parser"
)

func TestRunReturnsTheExpectedExitCodeForCLIArguments(t *testing.T) {
	tests := map[string]struct {
		args []string
		want int
	}{
		"rejects invalid weather":                       {args: []string{"-w", "99", "showdown_demo_files/player.txt", "showdown_demo_files/opponent.txt"}, want: 1},
		"rejects negative weather":                      {args: []string{"-w", "-1", "showdown_demo_files/player.txt", "showdown_demo_files/opponent.txt"}, want: 1},
		"rejects non-positive iterations":               {args: []string{"--iterations", "0", "showdown_demo_files/player.txt", "showdown_demo_files/opponent.txt"}, want: 1},
		"rejects negative iterations":                   {args: []string{"--iterations", "-5", "showdown_demo_files/player.txt", "showdown_demo_files/opponent.txt"}, want: 1},
		"rejects missing args":                          {args: []string{"--iterations", "1"}, want: 1},
		"rejects too many args":                         {args: []string{"showdown_demo_files/player.txt", "showdown_demo_files/opponent.txt", "extra.txt"}, want: 1},
		"rejects invalid showdown file":                 {args: []string{"--iterations", "1", "data/nonexistent.txt", "showdown_demo_files/opponent.txt"}, want: 1},
		"rejects misconfigured policy file":             {args: []string{"--policy-file", "missing.json", "--iterations", "1"}, want: 1},
		"rejects policy file combined with party files": {args: []string{"--policy-file", "missing.json", "showdown_demo_files/player.txt", "showdown_demo_files/opponent.txt"}, want: 1},
		"accepts valid CLI and returns zero":            {args: []string{"--iterations", "1", "showdown_demo_files/player.txt", "showdown_demo_files/opponent.txt"}, want: 0},
		"accepts a help request and returns zero":       {args: []string{"-h"}, want: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := run(tc.args); got != tc.want {
				t.Fatalf("run(%v) = %d, want %d", tc.args, got, tc.want)
			}
		})
	}
}

func TestRunAcceptsVerboseFlag(t *testing.T) {
	oldVerbose := *verbose
	*verbose = false
	defer func() { *verbose = oldVerbose }()

	code := run([]string{"-v", "--iterations", "1", "showdown_demo_files/player.txt", "showdown_demo_files/opponent.txt"})
	if code != 0 {
		t.Fatalf("expected zero exit code with verbose flag, got %d", code)
	}
	if !*verbose {
		t.Fatal("expected verbose flag to be enabled when parsing -v")
	}
}

func TestRunHelpDoesNotReturnError(t *testing.T) {
	if code := run([]string{"-h"}); code != 0 {
		t.Fatalf("expected zero exit code for help request, got %d", code)
	}
}

func TestRunLoadsPartiesEmbeddedInPolicyWhenNoInputFilesAreGiven(t *testing.T) {
	dir := t.TempDir()

	playerContent, err := os.ReadFile("showdown_demo_files/player.txt")
	if err != nil {
		t.Fatalf("read player fixture: %v", err)
	}
	opponentContent, err := os.ReadFile("showdown_demo_files/opponent.txt")
	if err != nil {
		t.Fatalf("read opponent fixture: %v", err)
	}

	withParties := filepath.Join(dir, "with_parties.json")
	writePolicyFixture(t, withParties, engine.SavedPolicy{
		PlayerParty:   string(playerContent),
		OpponentParty: string(opponentContent),
		Policy:        map[string][]string{},
		Scores:        map[string]map[string]float64{},
		Counts:        map[string]map[string]int{},
	})

	if code := run([]string{"--policy-file", withParties, "--iterations", "1"}); code != 0 {
		t.Fatalf("run(--policy-file, no input files) = %d, want 0", code)
	}

	withoutParties := filepath.Join(dir, "without_parties.json")
	writePolicyFixture(t, withoutParties, engine.SavedPolicy{
		Policy: map[string][]string{},
		Scores: map[string]map[string]float64{},
		Counts: map[string]map[string]int{},
	})

	if code := run([]string{"--policy-file", withoutParties, "--iterations", "1"}); code != 1 {
		t.Fatalf("run(--policy-file without embedded parties, no input files) = %d, want 1", code)
	}
}

func writePolicyFixture(t *testing.T, path string, policy engine.SavedPolicy) {
	t.Helper()
	data, err := json.Marshal(policy)
	if err != nil {
		t.Fatalf("marshal policy fixture: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write policy fixture: %v", err)
	}
}

func TestParserEdgeCases(t *testing.T) {
	tests := map[string]struct {
		input string
		want  bool
	}{
		"missing level line":     {input: "Horsea\nModest Nature\nAbility: Swift Swim\n- Bubble Beam\n", want: false},
		"invalid nature":         {input: "Horsea\nLevel: 17\nBad Nature\nAbility: Swift Swim\n- Bubble Beam\n", want: false},
		"invalid status":         {input: "Horsea\nLevel: 17\nModest Nature\nAbility: Swift Swim\nStatus> bad\n- Bubble Beam\n", want: false},
		"valid multi-move party": {input: "Horsea\nLevel: 17\nModest Nature\nAbility: Swift Swim\n- Bubble Beam\n- Twister\n", want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "party.txt")
			if err := os.WriteFile(path, []byte(tc.input), 0o644); err != nil {
				t.Fatalf("write temp file: %v", err)
			}

			parsed, err := parser.ReadShowdownFile(path)
			if tc.want {
				if err != nil {
					t.Fatalf("expected valid parser result, got error: %v", err)
				}
				if len(parsed) == 0 {
					t.Fatal("expected at least one parsed pokemon")
				}
				if parsed[0].Name != "Horsea" {
					t.Fatalf("unexpected parsed name: %q", parsed[0].Name)
				}
				return
			}
			if err == nil {
				t.Fatal("expected parse error for malformed input")
			}
		})
	}
}

func TestRunRejectsInvalidPolicyFileFormat(t *testing.T) {
	dir := t.TempDir()
	badPolicy := filepath.Join(dir, "bad_policy.json")
	if err := os.WriteFile(badPolicy, []byte("{invalid json}"), 0o644); err != nil {
		t.Fatalf("write bad policy file: %v", err)
	}

	code := run([]string{"--policy-file", badPolicy, "--iterations", "1"})
	if code != 1 {
		t.Fatalf("expected exit code 1 for invalid policy file, got %d", code)
	}
}

func TestPolicyCliEndToEndSmoke(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(cwd) }()

	cmdSave := exec.Command("go", "run", ".", "--player-learning-ai", "--iterations", "10", "showdown_demo_files/player.txt", "showdown_demo_files/opponent.txt")
	cmdSave.Dir = cwd
	output, err := cmdSave.CombinedOutput()
	if err != nil {
		t.Fatalf("save policy CLI failed: %v\n%s", err, output)
	}
	if !bytes.Contains(output, []byte("policy saved to")) {
		t.Fatalf("save CLI did not report policy saved:\n%s", output)
	}

	policyPath := filepath.Join(cwd, "policies", "player__vs__rnb_trainer_1.json")
	if _, err := os.Stat(policyPath); err != nil {
		t.Fatalf("saved policy file was not created at %s: %v", policyPath, err)
	}

	beforeReload, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("reading saved policy before reload failed: %v", err)
	}

	cmdLoad := exec.Command("go", "run", ".", "--policy-file", policyPath, "--iterations", "1")
	cmdLoad.Dir = cwd
	output, err = cmdLoad.CombinedOutput()
	if err != nil {
		t.Fatalf("load policy CLI failed: %v\n%s", err, output)
	}
	if !bytes.Contains(output, []byte("loaded policy from")) {
		t.Fatalf("load CLI did not report the loaded policy summary:\n%s", output)
	}

	afterReload, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("reading saved policy after reload failed: %v", err)
	}
	if !bytes.Equal(beforeReload, afterReload) {
		t.Fatalf("policy should remain identical before and after reload for the same input pair\nbefore=%s\nafter=%s", beforeReload, afterReload)
	}

	cmdLoadWithExtraArgs := exec.Command("go", "run", ".", "--policy-file", policyPath, "showdown_demo_files/player.txt", "showdown_demo_files/opponent.txt", "--iterations", "1")
	cmdLoadWithExtraArgs.Dir = cwd
	if output, err = cmdLoadWithExtraArgs.CombinedOutput(); err == nil {
		t.Fatalf("expected an error when passing party files alongside --policy-file:\n%s", output)
	}
}
