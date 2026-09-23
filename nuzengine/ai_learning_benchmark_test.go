package nuzengine

import (
	"os"
	"sync"
	"testing"

	"github.com/fred-bonn/nuz/nuzengine/internal/pokeapi"
)

func LearnSingle(playerPartyStr, opponentPartyStr string, weatherInt, iterations int) {
	cfg := &config{}
	runWorker(cfg, playerPartyStr, opponentPartyStr, weatherInt, iterations)
}

func LearnParallel(playerPartyStr, opponentPartyStr string, weatherInt, iterations int) {
	Verbose = false
	cfg := &config{
		client: pokeapi.NewClient(),
	}

	episodesPerWorker := iterations / monteCarloWorkers

	results := make(chan qMap, monteCarloWorkers)
	var wg sync.WaitGroup
	for range monteCarloWorkers {
		wg.Go(func() {
			q := runWorker(cfg, playerPartyStr, opponentPartyStr, weatherInt, episodesPerWorker)
			if q != nil {
				results <- q
			}
		})
	}
	wg.Wait()
	close(results)

	var workerMaps []qMap
	for q := range results {
		workerMaps = append(workerMaps, q)
	}
}

func BenchmarkLearnParallel(b *testing.B) {
	benchmarkLearn(b, LearnParallel)
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
