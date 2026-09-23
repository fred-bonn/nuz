package nuzengine

import (
	"fmt"
	"os"

	"github.com/fred-bonn/nuz/nuzengine/internal/pokeapi"
)

type battleState interface {
	execute() error
	reset()
	setError(error)
	gatherActions()
	getAllSlots() []*slot
	getOtherSlots(*slot) []*slot
	getOpponentSlot(*slot) *slot
	getPlayerTrainer() *trainer
	getPokemonSlot(*pokemon) *slot
	getActions() *actionQueue
	getWeather() weatherState
	setWeather(weatherState)
	getFieldEffects() map[fieldEffect]int
	key() string
}

func InitBattleState(battleStateInt int, playerPartyStr, opponentPartyStr string, aiInt, weatherInt int, policyFile string) (battleState, error) {
	cfg := &config{
		client: pokeapi.NewClient(),
	}

	if battleStateInt < 0 || battleStateInt > 1 {
		return nil, fmt.Errorf("invalid battle state type: %d", battleStateInt)
	}

	if aiInt < 0 || aiInt > 3 {
		return nil, fmt.Errorf("invalid AI type: %d", aiInt)
	}
	var playerAi ai
	if policyFile != "" {
		policy, err := loadPolicy(policyFile)
		if err != nil {
			return nil, fmt.Errorf("failed loading policy file: %w", err)
		}
		playerPartyStr = policy.PlayerParty
		opponentPartyStr = policy.OpponentParty
		weatherInt = policy.Weather
		playerAi = newPolicyAI(policy.Policy)
	} else {
		switch aiInt {
		case 0:
			playerAi = &rnbAi{}
		case 1:
			playerAi = newGuidedAI(os.Stdin, os.Stdout)
		case 2:
			playerAi = &randomAi{}
		}
	}

	if weatherInt < 0 || weatherInt > 4 {
		return nil, fmt.Errorf("invalid weather type: %d", weatherInt)
	}
	weather := weatherState(weatherInt)

	playerParty, err := cfg.validateInput(playerPartyStr)
	if err != nil {
		return nil, fmt.Errorf("failed validating player party: %s", err)
	}

	opponentParty, err := cfg.validateInput(opponentPartyStr)
	if err != nil {
		return nil, fmt.Errorf("failed validating opponent party: %s", err)
	}

	var battleState battleState

	switch battleStateInt {
	case 0:
		battleState = initSingleBattleState(
			trainer{
				ai:           playerAi,
				player:       true,
				fieldEffects: make(map[fieldEffect]int),
			},
			trainer{
				ai:           rnbAi{},
				fieldEffects: make(map[fieldEffect]int),
			},
			playerParty,
			opponentParty,
			weather,
		)
	}

	return battleState, nil
}

func Execute(bs battleState, iterations int) error {
	statistics := newBattleStatistics(bs)

	if err := bs.execute(); err != nil {
		return err
	}
	statistics.record()

	for i := iterations - 1; i > 0; i-- {
		bs.reset()
		if err := bs.execute(); err != nil {
			return err
		}
		statistics.record()
	}

	if iterations > 1 {
		statistics.print()
	}

	return nil
}

func injectReplaceAction(bs battleState, slot *slot, midTurn bool) {
	bs.getActions().queue.push(&replaceAction{
		oldSlot: slot,
		Trainer: slot.Trainer,
		midTurn: midTurn,
	})
	bs.getActions().sort(bs)
}

func resolveEndOfTurn(bs battleState) {
	for _, slot := range bs.getAllSlots() {
		// resolve end of return effects from ailments and statuses
		for _, ailment := range slot.mon.ailments {
			switch ailment.State {
			case burnAilment:
				takeResidualDamage(bs, slot, ailment.State.String(), 1, 16)
			case poisonAilment:
				takeResidualDamage(bs, slot, ailment.State.String(), 1, 8)
			case toxicAilment:
				ailment.Turns++
				takeResidualDamage(bs, slot, ailment.State.String(), ailment.Turns, 16)
			case trapAilment:
				ailment.Turns--
				takeResidualDamage(bs, slot, ailment.State.String(), 1, 8)
				if ailment.Turns <= 0 {
					vprintf("%s was freed", slot.mon.base.Name)
					delete(slot.mon.ailments, ailment.State)
				}
			case leechSeedAilment:
				vprintf("%s leeched health from %s", ailment.afflictedBy.mon.base.Name, slot.mon.base.Name)
				dmg := takeResidualDamage(bs, slot, ailment.State.String(), 1, 8)
				ailment.afflictedBy.mon.ChangeHpBy(dmg)
			case yawnAilment:
				ailment.Turns--
				if ailment.Turns == 0 {
					slot.mon.applyAilment(sleepAilment, nil, ailment.afflictedBy)
					delete(slot.mon.ailments, ailment.State)
				}
			}
		}

		// resolve end of turn effects of weather
		if w := bs.getWeather(); w != noneWeather {
			if w.affectsMon(slot.mon) {
				takeResidualDamage(bs, slot, w.String(), 1, 16)
			}
			w.activateMonAbility(bs, slot)
		}

		// reset protect counter if the slot was not protected this turn
		if !slot.protected {
			slot.protectTurns = 0
		} else {
			slot.protected = false
		}

		if slot.mon.ability == harvestAbility && roll(1, 2) && slot.mon.item.State.isBerry() {
			vprintf("%s harvested its %s", slot.mon.base.Name, slot.mon.item.String())
			slot.mon.item.Consumed = false
			slot.mon.checkItemTrigger(true, nil)
		} else if slot.mon.ability == speedBoostAbility && !slot.firstTurn {
			slot.mon.changeStatStageBy(speed, 1, false)
		}

		if slot.mon.item.State == leftovers {
			change := slot.mon.MaxHP() / 16
			vprintItem("%s restored %d health from leftovers", slot.mon.base.Name, change)
			slot.mon.ChangeHpBy(change)
		}

		slot.mon.laserFocus = false
	}
}

func takeResidualDamage(bs battleState, slot *slot, effect string, num, den int) int {
	if slot.mon.fainted {
		return 0
	}

	change := slot.mon.MaxHP() * num / den
	vprintf("%s took %d damage from %s", slot.mon.base.Name, change, effect)
	slot.mon.ChangeHpBy(-change)
	if slot.mon.hp <= 0 {
		slot.mon.fainted = true
		injectReplaceAction(bs, slot, false)
		vprintf("%s fainted!", slot.mon.base.Name)
	}
	return change
}

func resolveOnEntry(bs battleState) {
	for _, slot := range bs.getAllSlots() {
		if f, ok := onSwitchAbilities[slot.mon.ability]; ok {
			f(slot, bs, true)
		}
	}
}

func resetPokemonParty(party []*pokemon) {
	for _, mon := range party {
		mon.reset()
	}
}
