package engine

import (
	"math/rand"
)

type ailmentState int

const (
	paralysisAilment ailmentState = iota
	poisonAilment
	toxicAilment
	burnAilment
	freezeAilment
	sleepAilment
	infatuationAilment
	confusionAilment
	trapAilment
	boundAilment
	leechSeedAilment
	yawnAilment
	noneAilment
)

func (as ailmentState) isNonVolatileStatus() bool {
	return as >= paralysisAilment && as <= sleepAilment
}

func nonVolatileStatuses(yield func(ailmentState) bool) {
	for s := paralysisAilment; s <= sleepAilment; s++ {
		if !yield(s) {
			return
		}
	}
}

func volatileStatuses(yield func(ailmentState) bool) {
	for s := infatuationAilment; s < noneAilment; s++ {
		if !yield(s) {
			return
		}
	}
}

func StringToAilmentState(s string) ailmentState {
	switch s {
	case "paralysis":
		return paralysisAilment
	case "poison":
		return poisonAilment
	case "toxic":
		return toxicAilment
	case "burn":
		return burnAilment
	case "freeze":
		return freezeAilment
	case "sleep":
		return sleepAilment
	case "infatuation":
		return infatuationAilment
	case "confusion":
		return confusionAilment
	case "trap":
		return trapAilment
	case "bound":
		return boundAilment
	case "leech seed":
		return leechSeedAilment
	case "yawn":
		return yawnAilment
	case "none", "":
		return noneAilment
	default:
		elogf("warning: %s is not a valid ailment and was made into noneAilment", s)
		return noneAilment
	}
}

func (as ailmentState) String() string {
	switch as {
	case paralysisAilment:
		return "paralysis"
	case poisonAilment:
		return "poison"
	case toxicAilment:
		return "toxic"
	case burnAilment:
		return "burn"
	case freezeAilment:
		return "freeze"
	case sleepAilment:
		return "sleep"
	case infatuationAilment:
		return "infatuation"
	case confusionAilment:
		return "confusion"
	case trapAilment:
		return "trap"
	case boundAilment:
		return "bound"
	case leechSeedAilment:
		return "leech seed"
	case yawnAilment:
		return "yawn"
	default:
		elogf("warning: ailmenState.String(): ailmentState %d does not have a string version", as)
		return ""
	}
}

type ailment struct {
	State       ailmentState
	Turns       int
	afflictedBy *slot
}

func GenerateAilment(as ailmentState, afflictedBy *slot) *ailment {
	res := ailment{
		State:       as,
		afflictedBy: afflictedBy,
	}
	switch as {
	case sleepAilment:
		res.Turns = rand.Intn(3) + 1
	case confusionAilment:
		res.Turns = rand.Intn(4) + 1
	case yawnAilment:
		res.Turns = 2
	}
	return &res
}

func generateTrap(low, high int, mon *slot) *ailment {
	return &ailment{
		State:       trapAilment,
		Turns:       rand.Intn(high-low+1) + low,
		afflictedBy: mon,
	}
}
