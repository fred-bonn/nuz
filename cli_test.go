package main

import (
	"testing"
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
