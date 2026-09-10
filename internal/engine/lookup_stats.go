package engine

type statState int

const (
	hitPoints statState = iota
	attack
	defense
	specialAttack
	specialDefense
	Speed
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
	case Speed:
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
		return Speed
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
	"brave":   {attack, Speed},
	"bold":    {defense, attack},
	"docile":  {defense, defense},
	"impish":  {defense, specialAttack},
	"lax":     {defense, specialDefense},
	"relaxed": {defense, Speed},
	"modest":  {specialAttack, attack},
	"mild":    {specialAttack, defense},
	"bashful": {specialAttack, specialAttack},
	"rash":    {specialAttack, specialDefense},
	"quiet":   {specialAttack, Speed},
	"calm":    {specialDefense, attack},
	"gentle":  {specialDefense, defense},
	"careful": {specialDefense, Speed},
	"quirky":  {specialDefense, specialDefense},
	"sassy":   {specialDefense, Speed},
	"timid":   {Speed, attack},
	"hasty":   {Speed, defense},
	"jolly":   {Speed, specialAttack},
	"naive":   {Speed, specialDefense},
	"serious": {Speed, Speed},
}
