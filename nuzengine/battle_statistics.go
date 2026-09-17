package nuzengine

import (
	"log"
)

type battleStatistics struct {
	battleState      battleState
	battleCount      int
	winCount         int
	monSurvivalCount []int
}

func newBattleStatistics(bs battleState) *battleStatistics {
	return &battleStatistics{
		battleState:      bs,
		monSurvivalCount: make([]int, len(bs.getPlayerTrainer().pokemonParty)),
	}
}

func (bs *battleStatistics) record() {
	bs.battleCount++
	if bs.battleState.getPlayerTrainer().lost == false {
		bs.winCount++
	}

	for i, mon := range bs.battleState.getPlayerTrainer().pokemonParty {
		if !mon.fainted {
			bs.monSurvivalCount[i]++
		}
	}
}

func (bs *battleStatistics) print() {
	if bs.battleCount == 0 {
		return
	}

	winRate := float64(bs.winCount) * 100.0 / float64(bs.battleCount)
	log.Printf("player win rate: %.2f%% (%d/%d)", winRate, bs.winCount, bs.battleCount)
	for index, mon := range bs.battleState.getPlayerTrainer().pokemonParty {
		survivalRate := float64(bs.monSurvivalCount[index]) * 100.0 / float64(bs.battleCount)
		log.Printf("%s survival rate: %.2f%% (%d/%d)", mon.base.Name, survivalRate, bs.monSurvivalCount[index], bs.battleCount)
	}
}

func (bs *battleStatistics) partySurvivalRatio() float64 {
	if bs == nil || bs.battleCount == 0 || len(bs.monSurvivalCount) == 0 {
		return 0
	}

	alive := 0.0
	for _, survivors := range bs.monSurvivalCount {
		alive += float64(survivors) / float64(bs.battleCount)
	}
	return alive / float64(len(bs.monSurvivalCount))
}

func (bs *battleStatistics) AllPartySurvived() bool {
	if bs == nil || bs.battleCount == 0 || len(bs.monSurvivalCount) == 0 {
		return false
	}
	for _, survivors := range bs.monSurvivalCount {
		if survivors != bs.battleCount {
			return false
		}
	}
	return true
}

func (bs *battleStatistics) outcomeScore() float64 {
	if bs == nil || bs.battleCount == 0 {
		return 0
	}

	winRate := float64(bs.winCount) / float64(bs.battleCount)
	survivalRatio := bs.partySurvivalRatio()
	lossRate := 1.0 - winRate
	missingMembers := 0
	deadMembers := 0
	aliveMembers := 0
	for _, survivors := range bs.monSurvivalCount {
		if survivors == bs.battleCount {
			aliveMembers++
		}
		if survivors != bs.battleCount {
			missingMembers++
			deadMembers++
		}
	}
	aliveRatio := float64(aliveMembers) / float64(len(bs.monSurvivalCount))

	if bs.AllPartySurvived() {
		return 1000000.0 + 750000.0*winRate + 250000.0*survivalRatio
	}

	winBonus := 0.0
	if winRate > 0.0 {
		winBonus = 750000.0 * winRate
	}
	lossPenalty := 0.0
	if lossRate > 0.0 {
		lossPenalty = 900000.0 * lossRate
	}

	score := winBonus - lossPenalty
	score += aliveRatio * 200000.0
	score += survivalRatio * 120000.0
	score -= float64(deadMembers) * 500000.0
	score -= float64(missingMembers) * 250000.0
	if winRate > 0.5 && survivalRatio > 0.5 {
		score += 500000.0
	}
	if winRate == 0.0 {
		score -= 250000.0
	}
	if winRate == 1.0 {
		score += 500000.0
	}
	return score
}
