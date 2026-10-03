package nuzengine

import (
	"slices"
	"testing"
)

type statelessEntry = struct {
	count int
	value float64
}

func TestStatelessBestKeysCollectsAllKeysTiedAtBestScore(t *testing.T) {
	la := newLearningAIStateless()
	for _, key := range []string{"move:a", "move:b"} {
		la.sequence = []string{key}
		la.endEpisode(-10)
	}
	la.sequence = []string{"move:c"}
	la.endEpisode(-20)

	got := slices.Clone(la.bestKeys)
	slices.Sort(got)
	if !slices.Equal(got, []string{"move:a", "move:b"}) {
		t.Fatalf("bestKeys = %v, want [move:a move:b]", got)
	}
}

func TestStatelessBestKeysResetsOnHigherScore(t *testing.T) {
	la := newLearningAIStateless()
	la.sequence = []string{"move:a"}
	la.endEpisode(-10)
	la.sequence = []string{"move:b"}
	la.endEpisode(0)

	if !slices.Equal(la.bestKeys, []string{"move:b"}) {
		t.Fatalf("bestKeys = %v, want [move:b]", la.bestKeys)
	}
}

func TestStatelessBestKeysRecomputesWhenOnlyBestScoreFalls(t *testing.T) {
	la := newLearningAIStateless()
	la.sMap["move:best"] = statelessEntry{count: 1, value: 0}
	la.sMap["move:next"] = statelessEntry{count: 1, value: -10}
	la.resetBest("move:best", 0)
	la.sequence = []string{"move:best"}

	la.endEpisode(-40)

	if !slices.Equal(la.bestKeys, []string{"move:next"}) {
		t.Fatalf("bestKeys = %v, want [move:next]", la.bestKeys)
	}
}

func TestStatelessBestKeysDropsOneTiedKeyWhenItsScoreFalls(t *testing.T) {
	la := newLearningAIStateless()
	la.sMap["move:a"] = statelessEntry{count: 1, value: 0}
	la.sMap["move:b"] = statelessEntry{count: 1, value: 0}
	la.resetBest("move:a", 0)
	la.addBest("move:b")
	la.sequence = []string{"move:a"}

	la.endEpisode(-20)

	if !slices.Equal(la.bestKeys, []string{"move:b"}) {
		t.Fatalf("bestKeys = %v, want [move:b]", la.bestKeys)
	}
	if _, ok := la.bestIndex["move:a"]; ok {
		t.Fatal("move:a should no longer be indexed as best")
	}
}
