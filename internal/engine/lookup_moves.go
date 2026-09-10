package engine

var sleepMoves = map[string]struct{}{
	"dark void":     {},
	"grass whistle": {},
	"hypnosis":      {},
	"lovely kiss":   {},
	"sing":          {},
	"sleep powder":  {},
	"spore":         {},
	"yawn":          {},
}

func isSleepMove(move string) bool {
	_, ok := sleepMoves[move]
	return ok
}

var pivotMoves = map[string]struct{}{
	"u turn":           {},
	"volt switch":      {},
	"flip turn":        {},
	"parting shot":     {},
	"teleport":         {},
	"chilly reception": {},
	"baton pass":       {},
	"shed tail":        {},
}

func isPivotMove(move string) bool {
	_, ok := pivotMoves[move]
	return ok
}

var speedControlMoves = map[string]struct{}{
	"electroweb": {},
	"icy wind":   {},
	"low sweep":  {},
	"mud shot":   {},
	"rock tomb":  {},
	"bulldoze":   {},
	"glaciate":   {},
}

func isSpeedControlMove(move string) bool {
	_, ok := speedControlMoves[move]
	return ok
}

var offenseControlMoves = map[string]moveClass{
	"mystical fire":  SpecialClass,
	"skitter smack":  SpecialClass,
	"breaking swipe": physicalClass,
	"snarl":          SpecialClass,
	"struggle bug":   SpecialClass,
	"trop kick":      SpecialClass,
	"chilling water": physicalClass,
	"lunge":          physicalClass,
}

func isOffenseControlMove(move string) (moveClass, bool) {
	c, ok := offenseControlMoves[move]
	return c, ok
}

var protectMoves = map[string]struct{}{
	"protect":      {},
	"detect":       {},
	"kings shield": {},
}

func isProtectMove(move string) bool {
	_, ok := protectMoves[move]
	return ok
}

var multipleTurnMoves = map[string]struct{}{
	"bounce": {},
}

func isMultipleTurnMove(move string) bool {
	_, ok := multipleTurnMoves[move]
	return ok
}

var paralysisMoves = map[string]struct{}{
	"thunder wave": {},
	"glare":        {},
	"stun spore":   {},
	"nuzzle":       {},
}

func isParalysisMove(move string) bool {
	_, ok := paralysisMoves[move]
	return ok
}

var powderMoves = map[string]struct{}{
	"powder":        {},
	"spore":         {},
	"sleep powder":  {},
	"stun spore":    {},
	"poison powder": {},
	"rage powder":   {},
	"cotten spore":  {},
}

func isPowderMove(move string) bool {
	_, ok := powderMoves[move]
	return ok
}

var selfThawingMoves = map[string]struct{}{
	"flame wheel":     {},
	"sacred fire":     {},
	"flare blitz":     {},
	"fusion flare":    {},
	"scald":           {},
	"steam eruption":  {},
	"burn up":         {},
	"pyro ball":       {},
	"scorching sands": {},
}

func isSelfThawingMove(move string) bool {
	_, ok := selfThawingMoves[move]
	return ok
}
