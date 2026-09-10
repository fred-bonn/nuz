package engine

import "log"

type battleStatistics struct {
	winCount         int
	battleCount      int
	pokemonSurvivors []int
}

func newBattleStatistics(party []*Pokemon) battleStatistics {
	return battleStatistics{
		pokemonSurvivors: make([]int, len(party)),
	}
}

func (bs *battleStatistics) record(player *Trainer) {
	bs.battleCount++
	if !player.lost {
		bs.winCount++
	}

	for index, mon := range player.PokemonParty {
		if !mon.fainted {
			bs.pokemonSurvivors[index]++
		}
	}
}

func (bs *battleStatistics) partySurvivalRatio() float64 {
	if bs == nil || bs.battleCount == 0 || len(bs.pokemonSurvivors) == 0 {
		return 0
	}

	alive := 0.0
	for _, survivors := range bs.pokemonSurvivors {
		alive += float64(survivors) / float64(bs.battleCount)
	}
	return alive / float64(len(bs.pokemonSurvivors))
}

func (bs *battleStatistics) AllPartySurvived() bool {
	if bs == nil || bs.battleCount == 0 || len(bs.pokemonSurvivors) == 0 {
		return false
	}
	for _, survivors := range bs.pokemonSurvivors {
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
	for _, survivors := range bs.pokemonSurvivors {
		if survivors == bs.battleCount {
			aliveMembers++
		}
		if survivors != bs.battleCount {
			missingMembers++
			deadMembers++
		}
	}
	aliveRatio := float64(aliveMembers) / float64(len(bs.pokemonSurvivors))

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

func (bs *battleStatistics) print(party []*Pokemon) {
	if bs.battleCount == 0 {
		return
	}

	winRate := float64(bs.winCount) * 100.0 / float64(bs.battleCount)
	log.Printf("player win rate: %.2f%% (%d/%d)", winRate, bs.winCount, bs.battleCount)
	for index, mon := range party {
		survivalRate := float64(bs.pokemonSurvivors[index]) * 100.0 / float64(bs.battleCount)
		log.Printf("%s survival rate: %.2f%% (%d/%d)", mon.Base.Name, survivalRate, bs.pokemonSurvivors[index], bs.battleCount)
	}
}
