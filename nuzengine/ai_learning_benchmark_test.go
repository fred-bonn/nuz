package nuzengine

import (
	"os"
	"testing"
)

func LearnSingle(playerPartyStr, opponentPartyStr string, weatherInt, iterations int) {
	cfg := &config{}
	q := runWorker(cfg, playerPartyStr, opponentPartyStr, weatherInt, iterations)
	runPolicy(cfg, q, playerPartyStr, opponentPartyStr, weatherInt)
}

func BenchmarkLearn(b *testing.B) {
	benchmarkLearn(b, Learn)
}

func BenchmarkLearnSingle(b *testing.B) {
	benchmarkLearn(b, LearnSingle)
}

func benchmarkLearn(b *testing.B, learn func(string, string, int, int)) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	if err := os.Chdir(".."); err != nil {
		b.Fatal(err)
	}
	defer os.Chdir(workingDirectory)

	playerParty, err := os.ReadFile("showdown_demo_files/player.txt")
	if err != nil {
		b.Fatal(err)
	}
	opponentParty, err := os.ReadFile("showdown_demo_files/opponent.txt")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for range b.N {
		learn(string(playerParty), string(opponentParty), 0, 100000)
	}
}
