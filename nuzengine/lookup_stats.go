package nuzengine

import "fmt"

type statState int

const (
	hitPoints statState = iota
	attack
	defense
	specialAttack
	specialDefense
	speed
	accuracy
	evasion
	noStat
)

func (s statState) String() string {
	switch s {
	case hitPoints:
		return "hp"
	case attack:
		return "attack"
	case defense:
		return "defense"
	case specialAttack:
		return "special attack"
	case specialDefense:
		return "special defense"
	case speed:
		return "speed"
	case accuracy:
		return "accuracy"
	case evasion:
		return "evasion"
	default:
		elogf("warning: stats.String(): stat %d does not have a string version", s)
		return ""
	}
}

func stringToStat(s string) statState {
	switch s {
	case "hp":
		return hitPoints
	case "attack":
		return attack
	case "defense":
		return defense
	case "special-attack":
		return specialAttack
	case "special-defense":
		return specialDefense
	case "speed":
		return speed
	case "accuracy":
		return accuracy
	case "evasion":
		return evasion
	default:
		return noStat
	}
}

type nature struct {
	positive statState
	negative statState
}

var natureChart = map[string]nature{
	"hardy":   {attack, attack},
	"lonely":  {attack, defense},
	"adamant": {attack, specialAttack},
	"naughty": {attack, specialDefense},
	"brave":   {attack, speed},
	"bold":    {defense, attack},
	"docile":  {defense, defense},
	"impish":  {defense, specialAttack},
	"lax":     {defense, specialDefense},
	"relaxed": {defense, speed},
	"modest":  {specialAttack, attack},
	"mild":    {specialAttack, defense},
	"bashful": {specialAttack, specialAttack},
	"rash":    {specialAttack, specialDefense},
	"quiet":   {specialAttack, speed},
	"calm":    {specialDefense, attack},
	"gentle":  {specialDefense, defense},
	"careful": {specialDefense, speed},
	"quirky":  {specialDefense, specialDefense},
	"sassy":   {specialDefense, speed},
	"timid":   {speed, attack},
	"hasty":   {speed, defense},
	"jolly":   {speed, specialAttack},
	"naive":   {speed, specialDefense},
	"serious": {speed, speed},
}

func getNature(nat string) (nature, error) {
	res, ok := natureChart[nat]
	if !ok {
		return nature{}, fmt.Errorf("invalid nature: %s", nat)
	}

	return res, nil

}
