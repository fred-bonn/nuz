package nuzengine

import (
	"os"
	"sync"
	"testing"

	"github.com/fred-bonn/nuz/nuzengine/internal/pokeapi"
)

func LearnSingle(playerPartyStr, opponentPartyStr string, weatherInt, iterations int) {
	cfg := &config{}
	runWorker(cfg, playerPartyStr, opponentPartyStr, weatherInt, iterations, func(episode) {})
}

func LearnParallelAggregator(playerPartyStr, opponentPartyStr string, weatherInt, iterations int) {
	Verbose = false
	cfg := &config{
		client: pokeapi.NewClient(),
	}

	episodesPerWorker := iterations / monteCarloWorkers

	episodes := make(chan episode, monteCarloWorkers)
	merged := newQMap()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for ep := range episodes {
			merged.recordEpisode(ep.trajectory, ep.reward)
		}
	}()

	var wg sync.WaitGroup
	for range monteCarloWorkers {
		wg.Go(func() {
			runWorker(cfg, playerPartyStr, opponentPartyStr, weatherInt, episodesPerWorker, func(ep episode) {
				episodes <- ep
			})
		})
	}
	wg.Wait()
	close(episodes)
	<-done
}

// LearnParallelMerge is the previous design: each worker builds its own qMap, merged after all finish.
func LearnParallelMerge(playerPartyStr, opponentPartyStr string, weatherInt, iterations int) {
	Verbose = false
	cfg := &config{
		client: pokeapi.NewClient(),
	}

	episodesPerWorker := iterations / monteCarloWorkers

	results := make(chan qMap, monteCarloWorkers)
	var wg sync.WaitGroup
	for range monteCarloWorkers {
		wg.Go(func() {
			q := newQMap()
			runWorker(cfg, playerPartyStr, opponentPartyStr, weatherInt, episodesPerWorker, func(ep episode) {
				q.recordEpisode(ep.trajectory, ep.reward)
			})
			results <- q
		})
	}
	wg.Wait()
	close(results)

	var workerMaps []qMap
	for q := range results {
		workerMaps = append(workerMaps, q)
	}
	mergeQMaps(workerMaps)
}

func mergeQMaps(maps []qMap) qMap {
	merged := newQMap()
	for _, m := range maps {
		for state, actions := range m {
			dst, ok := merged[state]
			if !ok {
				dst = make(map[string]*qEntry)
				merged[state] = dst
			}
			for action, entry := range actions {
				existing, ok := dst[action]
				if !ok {
					dst[action] = &qEntry{Value: entry.Value, Count: entry.Count}
					continue
				}
				totalCount := existing.Count + entry.Count
				existing.Value = (existing.Value*float64(existing.Count) + entry.Value*float64(entry.Count)) / float64(totalCount)
				existing.Count = totalCount
			}
		}
	}
	return merged
}

func BenchmarkLearnParallelAggregator(b *testing.B) {
	benchmarkLearn(b, LearnParallelAggregator)
}

func BenchmarkLearnParallelMerge(b *testing.B) {
	benchmarkLearn(b, LearnParallelMerge)
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
