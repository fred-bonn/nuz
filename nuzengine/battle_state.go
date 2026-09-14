package nuzengine

import (
	"fmt"
	"log"
	"os"

	"github.com/fred-bonn/nuz/nuzengine/internal/pokeapi"
)

type battleState interface {
	execute() error
	reset() error
	setError(error)
	gatherActions()
	getAllSlots() []*slot
	getOtherSlots(slot *slot) []*slot
	getOpponentSlot(slot *slot) *slot
	getPlayerTrainer() *trainer
	getActions() *actionQueue
	getWeather() weatherState
	setWeather(weatherState)
	getFieldEffects() map[fieldEffect]int
	getStatistics() *battleStatistics
	recordStatistics()
	printStatistics()
}

func InitBattleState(battleStateInt int, playerPartyStr, opponentPartyStr string, aiInt, weatherInt int, policyFile string) (battleState, error) {
	cfg := &config{
		client: pokeapi.NewClient(),
	}

	if battleStateInt < 0 || battleStateInt > 1 {
		return nil, fmt.Errorf("invalid battle state type: %d", battleStateInt)
	}

	if weatherInt < 0 || weatherInt > 4 {
		return nil, fmt.Errorf("invalid weather type: %d", weatherInt)
	}
	weather := weatherState(weatherInt)

	if aiInt < 0 || aiInt > 3 {
		return nil, fmt.Errorf("invalid AI type: %d", aiInt)
	}
	var playerAi ai
	if policyFile != "" {
		policy, err := loadPolicyFromDisk(policyFile)
		if err != nil {
			return nil, fmt.Errorf("failed loading policy '%s': %s", policyFile, err)
		}
		playerPartyStr = policy.PlayerParty
		opponentPartyStr = policy.OpponentParty
		playerAi = newStaticPolicyAIFromPolicy(policy)
	} else {
		switch aiInt {
		case 0:
			playerAi = rnbAi{}
		case 1:
			learningAi := newLearningAI()
			learningAi.playerShowdown = playerPartyStr
			learningAi.opponentShowdown = opponentPartyStr
			playerAi = learningAi
		case 2:
			playerAi = newGuidedAI(os.Stdin, os.Stdout)
		case 3:
			playerAi = randomAi{}
		}
	}

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
				AI:           playerAi,
				Player:       true,
				FieldEffects: make(map[fieldEffect]int),
			},
			trainer{
				AI:           rnbAi{},
				FieldEffects: make(map[fieldEffect]int),
			},
			playerParty,
			opponentParty,
			weather,
		)
	}

	return battleState, nil
}

func Execute(bs battleState, iterations int) error {
	var learning *learningAi
	if ai, ok := bs.getPlayerTrainer().AI.(*learningAi); ok {
		learning = ai
	}

	for range iterations {
		if err := bs.reset(); err != nil {
			return err
		}

		if err := bs.execute(); err != nil {
			return err
		}

		bs.recordStatistics()

		if learning != nil {
			learning.RecordBattleOutcome(bs.getStatistics())
		}
	}

	if learning != nil {
		if err := learning.savePolicyToDisk(); err != nil {
			log.Printf("error: failed saving policy: %s", err)
		} else {
			log.Printf("policy saved to policites/policy.json")
		}

	}

	if iterations > 1 {
		bs.printStatistics()
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
		for _, ailment := range slot.mon.Ailments {
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
					vprintf("%s was freed", slot.mon.Base.Name)
					delete(slot.mon.Ailments, ailment.State)
				}
			case leechSeedAilment:
				vprintf("%s leeched health from %s", ailment.afflictedBy.mon.Base.Name, slot.mon.Base.Name)
				dmg := takeResidualDamage(bs, slot, ailment.State.String(), 1, 8)
				ailment.afflictedBy.mon.ChangeHpBy(dmg)
			case yawnAilment:
				ailment.Turns--
				if ailment.Turns == 0 {
					slot.mon.applyAilment(sleepAilment, nil, ailment.afflictedBy)
					delete(slot.mon.Ailments, ailment.State)
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

		if slot.mon.Ability == harvestAbility && roll(1, 2) && slot.mon.Item.State.isBerry() {
			vprintf("%s harvested its %s", slot.mon.Base.Name, slot.mon.Item.String())
			slot.mon.Item.Consumed = false
			slot.mon.checkItemTrigger(true, nil)
		} else if slot.mon.Ability == speedBoostAbility && !slot.firstTurn {
			slot.mon.changeStatStageBy(Speed, 1, false)
		}

		if slot.mon.Item.State == leftovers {
			change := slot.mon.MaxHP() / 16
			vprintItem("%s restored %d health from leftovers", slot.mon.Base.Name, change)
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
	vprintf("%s took %d damage from %s", slot.mon.Base.Name, change, effect)
	slot.mon.ChangeHpBy(-change)
	if slot.mon.HP <= 0 {
		slot.mon.fainted = true
		injectReplaceAction(bs, slot, false)
		vprintf("%s fainted!", slot.mon.Base.Name)
	}
	return change
}

func resolveOnEntry(bs battleState) {
	for _, slot := range bs.getAllSlots() {
		if f, ok := onSwitchAbilities[slot.mon.Ability]; ok {
			f(slot, bs, true)
		}
	}
}
