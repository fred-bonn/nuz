package engine

type moveBalance struct {
	power        *int
	accuracy     *int
	pp           *int
	effectChance *int
	pokemonType  *pokemonType
}

func (mb moveBalance) Apply(m *Move) {
	if mb.power != nil {
		m.Power = *mb.power
	}
	if mb.accuracy != nil {
		m.Accuracy = *mb.accuracy
	}
	if mb.pp != nil {
		m.PP = *mb.pp
	}
	if mb.effectChance != nil {
		m.StatChance = *mb.effectChance
	}
	if mb.pokemonType != nil {
		m.Type = *mb.pokemonType
	}
}

var MoveBalanceMap = map[string]*moveBalance{
	"absorb": {
		power: new(40),
	},
	"aeroblast": {
		accuracy: new(100),
	},
	"air-cutter": {
		accuracy: new(100),
	},
	"air-slash": {
		accuracy: new(100),
	},
	"aqua-tail": {
		accuracy: new(95),
	},
	"astonish": {
		power: new(40),
	},
	"baby-doll-eyes": {
		pp: new(10),
	},
	"barrage": {
		accuracy: new(100),
	},
	"belch": {
		accuracy: new(100),
	},
	"bind": {
		accuracy: new(100),
	},
	"blaze-kick": {
		accuracy: new(100),
	},
	"blizzard": {
		accuracy: new(80),
	},
	"blue-flare": {
		accuracy: new(90),
	},
	"bolt-strike": {
		accuracy: new(90),
	},
	"bone-club": {
		accuracy: new(100),
	},
	"bonemerang": {
		accuracy: new(100),
	},
	"bounce": {
		accuracy: new(95),
	},
	"captivate": {
		pp: new(5),
	},
	"charge-beam": {
		power:        new(40),
		accuracy:     new(100),
		effectChance: new(100),
	},
	"charm": {
		pp: new(5),
	},
	"circle-throw": {
		accuracy: new(95),
	},
	"clamp": {
		accuracy: new(100),
	},
	"comet-punch": {
		accuracy: new(90),
	},
	"confide": {
		pp: new(10),
	},
	"covet": {
		pokemonType: new(fairyType),
	},
	"crabhammer": {
		accuracy: new(100),
	},
	"cross-chop": {
		accuracy: new(90),
	},
	"cut": {
		accuracy: new(100),
	},
	"dark-void": {
		accuracy: new(80),
	},
	"diamond-storm": {
		accuracy: new(100),
	},
	"double-hit": {
		accuracy: new(100),
	},
	"double-slap": {
		accuracy: new(100),
	},
	"draco-meteor": {
		accuracy: new(100),
	},
	"dragon-rush": {
		accuracy: new(85),
	},
	"dragon-tail": {
		accuracy: new(95),
	},
	"drill-run": {
		accuracy: new(100),
	},
	"dual-chop": {
		accuracy: new(100),
	},
	"dual-wingbeat": {
		accuracy: new(100),
	},
	"eerie-impulse": {
		pp: new(5),
	},
	"electroweb": {
		accuracy: new(100),
	},
	"fake-out": {
		pp: new(5),
	},
	"fake-tears": {
		pp: new(5),
	},
	"feather-dance": {
		pp: new(5),
	},
	"fire-fang": {
		accuracy: new(100),
	},
	"fire-spin": {
		accuracy: new(100),
	},
	"flash": {
		accuracy: new(70),
	},
	"fleur-cannon": {
		accuracy: new(100),
	},
	"fly": {
		accuracy: new(100),
	},
	"flying-press": {
		accuracy: new(100),
	},
	"focus-blast": {
		accuracy: new(80),
	},
	"freeze-shock": {
		accuracy: new(100),
	},
	"frenzy-plant": {
		accuracy: new(100),
	},
	"frost-breath": {
		accuracy: new(100),
	},
	"frustration": {
		power: new(102),
	},
	"fury-attack": {
		accuracy: new(100),
	},
	"fury-swipes": {
		accuracy: new(90),
	},
	"gear-grind": {
		accuracy: new(100),
	},
	"giga-impact": {
		accuracy: new(100),
	},
	"glaciate": {
		accuracy: new(100),
	},
	"grass-whistle": {
		accuracy: new(70),
	},
	"growl": {
		pp: new(10),
	},
	"gunk-shot": {
		accuracy: new(85),
	},
	"hammer-arm": {
		accuracy: new(100),
	},
	"harden": {
		pp: new(5),
	},
	"head-smash": {
		accuracy: new(85),
	},
	"heat-wave": {
		accuracy: new(100),
	},
	"high-horsepower": {
		accuracy: new(100),
	},
	"hurricane": {
		accuracy: new(80),
	},
	"hydro-cannon": {
		accuracy: new(100),
	},
	"hydro-pump": {
		accuracy: new(85),
	},
	"hyper-beam": {
		accuracy: new(100),
	},
	"hyper-fang": {
		accuracy: new(100),
	},
	"hypnosis": {
		accuracy: new(70),
	},
	"ice-burn": {
		accuracy: new(100),
	},
	"ice-fang": {
		accuracy: new(100),
	},
	"ice-hammer": {
		accuracy: new(100),
	},
	"icicle-crash": {
		accuracy: new(100),
	},
	"icy-wind": {
		accuracy: new(100),
	},
	"iron-tail": {
		accuracy: new(85),
	},
	"kinesis": {
		accuracy: new(100),
	},
	"leaf-storm": {
		accuracy: new(100),
	},
	"leaf-tornado": {
		accuracy:     new(100),
		effectChance: new(30),
	},
	"leech-seed": {
		accuracy: new(100),
	},
	"leer": {
		pp: new(10),
	},
	"lick": {
		power: new(40),
	},
	"light-of-ruin": {
		accuracy: new(100),
	},
	"lovely-kiss": {
		accuracy: new(80),
	},
	"magma-storm": {
		accuracy: new(90),
	},
	"mega-drain": {
		power: new(60),
	},
	"mega-kick": {
		accuracy: new(85),
	},
	"mega-punch": {
		accuracy: new(100),
	},
	"megahorn": {
		accuracy: new(90),
	},
	"metal-claw": {
		accuracy: new(100),
	},
	"metal-sound": {
		pp:       new(5),
		accuracy: new(100),
	},
	"meteor-beam": {
		accuracy: new(100),
	},
	"meteor-mash": {
		accuracy: new(100),
	},
	"mirror-shot": {
		accuracy:     new(100),
		effectChance: new(20),
	},
	"misty-explosion": {
		power: new(200),
	},
	"mud-bomb": {
		accuracy:     new(100),
		effectChance: new(20),
	},
	"muddy-water": {
		accuracy: new(95),
	},
	"natures-madness": {
		accuracy: new(100),
	},
	"night-daze": {
		accuracy:     new(100),
		effectChance: new(30),
	},
	"noble-roar": {
		pp: new(10),
	},
	"octazooka": {
		power:        new(80),
		accuracy:     new(100),
		effectChance: new(30),
	},
	"origin-pulse": {
		accuracy: new(100),
	},
	"overheat": {
		accuracy: new(100),
	},
	"pin-missile": {
		accuracy: new(100),
	},
	"play-nice": {
		pp: new(10),
	},
	"play-rough": {
		accuracy: new(100),
	},
	"poison-powder": {
		accuracy: new(90),
	},
	"power-whip": {
		accuracy: new(90),
	},
	"precipice-blades": {
		accuracy: new(100),
	},
	"psycho-boost": {
		accuracy: new(100),
	},
	"razor-leaf": {
		accuracy: new(100),
	},
	"razor-shell": {
		accuracy: new(100),
	},
	"return": {
		power: new(102),
	},
	"roar-of-time": {
		accuracy: new(100),
	},
	"rock-blast": {
		accuracy: new(100),
	},
	"rock-climb": {
		accuracy: new(95),
	},
	"rock-slide": {
		accuracy: new(100),
	},
	"rock-smash": {
		effectChance: new(100),
	},
	"rock-throw": {
		accuracy: new(100),
	},
	"rock-wrecker": {
		accuracy: new(100),
	},
	"rolling-kick": {
		accuracy: new(100),
	},
	"roost": {
		pp: new(5),
	},
	"sacred-fire": {
		accuracy: new(100),
	},
	"sand-tomb": {
		accuracy: new(100),
	},
	"sand-attack": {
		pp: new(5),
	},
	"scale-shot": {
		accuracy: new(100),
	},
	"screech": {
		pp:       new(5),
		accuracy: new(100),
	},
	"seed-flare": {
		accuracy: new(90),
	},
	"sing": {
		accuracy: new(70),
	},
	"skitter-smack": {
		accuracy: new(100),
	},
	"sky-attack": {
		accuracy: new(100),
	},
	"sky-uppercut": {
		accuracy: new(100),
	},
	"slam": {
		accuracy: new(90),
	},
	"sleep-powder": {
		accuracy: new(80),
	},
	"smog": {
		accuracy: new(90),
	},
	"snarl": {
		pp:       new(10),
		accuracy: new(100),
	},
	"sonic-boom": {
		accuracy: new(100),
	},
	"spacial-rend": {
		accuracy: new(100),
	},
	"steam-eruption": {
		accuracy: new(100),
	},
	"steel-beam": {
		accuracy: new(100),
	},
	"steel-wing": {
		accuracy: new(100),
	},
	"stone-edge": {
		accuracy: new(85),
	},
	"strange-steam": {
		accuracy: new(100),
	},
	"struggle-bug": {
		pp: new(10),
	},
	"stun-spore": {
		accuracy: new(90),
	},
	"submission": {
		accuracy: new(100),
	},
	"super-fang": {
		accuracy:    new(100),
		pokemonType: new(darkType),
	},
	"supersonic": {
		accuracy: new(70),
	},
	"swagger": {
		accuracy: new(90),
	},
	"sweet-kiss": {
		accuracy: new(80),
	},
	"tail-slap": {
		accuracy: new(100),
	},
	"take-down": {
		accuracy: new(100),
	},
	"tearful-look": {
		pp: new(10),
	},
	"thunder": {
		accuracy: new(80),
	},
	"thunder-cage": {
		accuracy: new(100),
	},
	"thunder-fang": {
		accuracy: new(100),
	},
	"tickle": {
		pp: new(10),
	},
	"whirlpool": {
		accuracy: new(100),
	},
	"wrap": {
		accuracy: new(100),
	},
	"zen-headbutt": {
		accuracy: new(100),
	},
}
