package nuzengine

import "testing"

func TestStatelessBestKeyPrefersLowerCountOnEqualScore(t *testing.T) {
	la := newLearningAIStateless()
	la.sMap["move:older"] = struct {
		count int
		value float64
	}{count: 2, value: -10}
	la.sMap["move:newer"] = struct {
		count int
		value float64
	}{count: 0}
	la.bestKey = "move:older"
	la.sequence = []string{"move:newer"}

	la.endEpisode(-10)

	if la.bestKey != "move:newer" {
		t.Fatalf("bestKey = %q, want move:newer", la.bestKey)
	}
}

func TestStatelessBestKeyRecomputesWhenCurrentBestScoreFalls(t *testing.T) {
	la := newLearningAIStateless()
	la.sMap["move:best"] = struct {
		count int
		value float64
	}{count: 1, value: 0}
	la.sMap["move:next"] = struct {
		count int
		value float64
	}{count: 1, value: -10}
	la.bestKey = "move:best"
	la.sequence = []string{"move:best"}

	la.endEpisode(-20)

	if la.bestKey != "move:next" {
		t.Fatalf("bestKey = %q, want move:next", la.bestKey)
	}
}
