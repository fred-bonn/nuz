package engine

import (
	"math/rand"
)

type abilityState int

const (
	NoneAbility abilityState = iota
	insomniaAbility
	vitalSpiritAbility
	sweetVeilAbility
	GluttonyAbility
	innerFocusAbility
	shieldDustAbility
	overcoatAbility
	cheekPouchAbility
	sturdyAbility
	harvestAbility
	speedBoostAbility
	traceAbility
	unnerveAbility
	intimidateAbility
	regeneratorAbility
	naturalCureAbility
	drizzleAbility
	droughtAbility
	snowWarningAbility
	sandStreamAbility
	pranksterAbility
	earlyBirdAbility
	stickyHoldAbility
	liquidOozeAbility
	cottenDownAbility
	waterCompactionAbility
	roughSkinAbility
	ironBarbsAbility
	cuteCharmAbility
	flameBodyAbility
	poisonPointAbility
	effectSporeAbility
	poisonTouchAbility
	magicGuardAbility
	sandVeilAbility
	sandRushAbility
	sandForceAbility
	iceBodyAbility
	snowCloakAbility
	raindDishAbility
	drySkinAbility
	hydrationAbility
	solarPowerAbility
	serenceGraceAbility
	keenEyeAbility
	clearSmokeAbility
	clearBodyAbility
	levitateAbility
	synchronizeAbility
	obliviousAbility
	immunityAbility
	corrosionAbility
	limberAbility
	waterVeilAbility
	slushRushAbility
	chlorophyllAbility
	swiftSwimAbility
	unburdenAbility
	battleArmorAbility
	shellArmorAbility
	magmaArmorAbility
	moldBreakerAbility
	mercilessAbility
	superLuckAbility
	sniperAbility
	compoundEyesAbility
	hustleAbility
	noGuardAbility
	technicianAbility
	overgrowAbility
	blazeAbility
	torrentAbility
	swarmAbility
	flashFireAbility
	waterAbsorbAbility
	stormDrainAbility
	voltAbsorbAbility
	lightningRodAbility
	motorDriveAbility
	sapSipperAbility
	aerilateAbility
	pixilateAbility
	galvanizeAbility
	refrigerateAbility
	normalizeAbility
)

func (a abilityState) String() string {
	switch a {
	case insomniaAbility:
		return "insomnia"
	case vitalSpiritAbility:
		return "vital spirit"
	case sweetVeilAbility:
		return "sweet veil"
	case GluttonyAbility:
		return "gluttony"
	case innerFocusAbility:
		return "inner focus"
	case shieldDustAbility:
		return "shield dust"
	case overcoatAbility:
		return "overcoat"
	case cheekPouchAbility:
		return "cheek pouch"
	case sturdyAbility:
		return "sturdy"
	case harvestAbility:
		return "harvest"
	case speedBoostAbility:
		return "speed boost"
	case traceAbility:
		return "trace"
	case unnerveAbility:
		return "unnerve"
	case intimidateAbility:
		return "intimidate"
	case regeneratorAbility:
		return "regenerator"
	case naturalCureAbility:
		return "natural-cure"
	case drizzleAbility:
		return "drizzle"
	case droughtAbility:
		return "drought"
	case snowWarningAbility:
		return "snow warning"
	case sandStreamAbility:
		return "sand stream"
	case pranksterAbility:
		return "prankster"
	case earlyBirdAbility:
		return "early bird"
	case stickyHoldAbility:
		return "sticky hold"
	case liquidOozeAbility:
		return "liquid ooze"
	case cottenDownAbility:
		return "cotten down"
	case waterCompactionAbility:
		return "water compaction"
	case roughSkinAbility:
		return "rought skin"
	case ironBarbsAbility:
		return "iron barbs"
	case cuteCharmAbility:
		return "cute charm"
	case flameBodyAbility:
		return "flame body"
	case poisonPointAbility:
		return "poison point"
	case effectSporeAbility:
		return "effect spore"
	case poisonTouchAbility:
		return "poison touch"
	case magicGuardAbility:
		return "magic guard"
	case sandVeilAbility:
		return "sand veil"
	case sandRushAbility:
		return "sand rush"
	case sandForceAbility:
		return "sand force"
	case iceBodyAbility:
		return "ice body"
	case snowCloakAbility:
		return "snow cloak"
	case raindDishAbility:
		return "raid dish"
	case drySkinAbility:
		return "dry skin"
	case hydrationAbility:
		return "hydration"
	case solarPowerAbility:
		return "solar power"
	case serenceGraceAbility:
		return "serence grace"
	case keenEyeAbility:
		return "keen eye"
	case clearSmokeAbility:
		return "clear smoke"
	case clearBodyAbility:
		return "clear body"
	case levitateAbility:
		return "levitate"
	case synchronizeAbility:
		return "synchronize"
	case obliviousAbility:
		return "oblivious"
	case immunityAbility:
		return "immunity"
	case corrosionAbility:
		return "corrosion"
	case limberAbility:
		return "limber"
	case waterVeilAbility:
		return "water veil"
	case slushRushAbility:
		return "slush rush"
	case chlorophyllAbility:
		return "chlorophyll"
	case swiftSwimAbility:
		return "swift swim"
	case unburdenAbility:
		return "unburden"
	case battleArmorAbility:
		return "battle armor"
	case shellArmorAbility:
		return "shell armor"
	case magmaArmorAbility:
		return "magma armor"
	case moldBreakerAbility:
		return "mold breaker"
	case mercilessAbility:
		return "merciless"
	case superLuckAbility:
		return "super kuck"
	case sniperAbility:
		return "sniper"
	case compoundEyesAbility:
		return "compound eyes"
	case hustleAbility:
		return "hustle"
	case noGuardAbility:
		return "no guard"
	case technicianAbility:
		return "technician"
	case overgrowAbility:
		return "overgrow"
	case blazeAbility:
		return "blaze"
	case torrentAbility:
		return "torrent"
	case swarmAbility:
		return "swarm"
	case flashFireAbility:
		return "flash fire"
	case waterAbsorbAbility:
		return "water absorb"
	case stormDrainAbility:
		return "storm drain"
	case voltAbsorbAbility:
		return "volt absorb"
	case lightningRodAbility:
		return "lightning rod"
	case motorDriveAbility:
		return "motor drive"
	case sapSipperAbility:
		return "sap sipper"
	case aerilateAbility:
		return "aerilate"
	case pixilateAbility:
		return "pixilate"
	case galvanizeAbility:
		return "galvanize"
	case refrigerateAbility:
		return "refrigerate"
	case normalizeAbility:
		return "normalize"
	default:
		elogf("warning: ability.String(): ability %d does not have a string version", a)
		return ""
	}
}

func StringToAbility(s string) abilityState {
	switch s {
	case "insomnia":
		return insomniaAbility
	case "vital spirit":
		return vitalSpiritAbility
	case "sweet veil":
		return sweetVeilAbility
	case "gluttony":
		return GluttonyAbility
	case "inner focus":
		return innerFocusAbility
	case "shield dust":
		return shieldDustAbility
	case "overcoat":
		return overcoatAbility
	case "cheek pouch":
		return cheekPouchAbility
	case "sturdy":
		return sturdyAbility
	case "harvest":
		return harvestAbility
	case "speed boost":
		return speedBoostAbility
	case "trace":
		return traceAbility
	case "unnerve":
		return unnerveAbility
	case "intimidate":
		return intimidateAbility
	case "regenerator":
		return regeneratorAbility
	case "natural cure":
		return naturalCureAbility
	case "drizzle":
		return drizzleAbility
	case "drought":
		return droughtAbility
	case "snow warning":
		return snowWarningAbility
	case "sand stream":
		return sandStreamAbility
	case "prankster":
		return pranksterAbility
	case "early bird":
		return earlyBirdAbility
	case "sticky hold":
		return stickyHoldAbility
	case "liquid ooze":
		return liquidOozeAbility
	case "cotten down":
		return cottenDownAbility
	case "water compaction":
		return waterCompactionAbility
	case "rough skin":
		return roughSkinAbility
	case "iron barbs":
		return ironBarbsAbility
	case "cute charm":
		return cuteCharmAbility
	case "flame body":
		return flameBodyAbility
	case "poison point":
		return poisonPointAbility
	case "effect spore":
		return effectSporeAbility
	case "poison touch":
		return poisonTouchAbility
	case "magic guard":
		return magicGuardAbility
	case "sand veil":
		return sandVeilAbility
	case "sand rush":
		return sandRushAbility
	case "sand force":
		return sandForceAbility
	case "ice body":
		return iceBodyAbility
	case "snow cloak":
		return snowCloakAbility
	case "rain dish":
		return raindDishAbility
	case "dry skin":
		return drySkinAbility
	case "hydration":
		return hydrationAbility
	case "solar power":
		return solarPowerAbility
	case "serene grace":
		return serenceGraceAbility
	case "keen eye":
		return keenEyeAbility
	case "clear smoke":
		return clearSmokeAbility
	case "clear body":
		return clearBodyAbility
	case "levitate":
		return levitateAbility
	case "synchronize":
		return synchronizeAbility
	case "oblivious":
		return obliviousAbility
	case "immunity":
		return immunityAbility
	case "corrosion":
		return corrosionAbility
	case "limber":
		return limberAbility
	case "water veil":
		return waterVeilAbility
	case "slush rush":
		return slushRushAbility
	case "chlorophyll":
		return chlorophyllAbility
	case "swift swim":
		return swiftSwimAbility
	case "unburden":
		return unburdenAbility
	case "battle armor":
		return battleArmorAbility
	case "shell armor":
		return shellArmorAbility
	case "magma armor":
		return magmaArmorAbility
	case "mold breaker":
		return moldBreakerAbility
	case "merciless":
		return mercilessAbility
	case "super luck":
		return superLuckAbility
	case "sniper":
		return sniperAbility
	case "compound eyes":
		return compoundEyesAbility
	case "hustle":
		return hustleAbility
	case "no guard":
		return noGuardAbility
	case "technician":
		return technicianAbility
	case "overgrow":
		return overgrowAbility
	case "blaze":
		return blazeAbility
	case "torrent":
		return torrentAbility
	case "swarm":
		return swarmAbility
	case "flash fire":
		return flashFireAbility
	case "water absorb":
		return waterAbsorbAbility
	case "storm drain":
		return stormDrainAbility
	case "volt absorb":
		return voltAbsorbAbility
	case "lightning rod":
		return lightningRodAbility
	case "motor drive":
		return motorDriveAbility
	case "sap sipper":
		return sapSipperAbility
	case "aerilate":
		return aerilateAbility
	case "pixilate":
		return pixilateAbility
	case "galvanize":
		return galvanizeAbility
	case "refrigerate":
		return refrigerateAbility
	case "normalize":
		return normalizeAbility
	default:
		return NoneAbility
	}
}

func (a abilityState) blocksCrits() bool {
	return a >= battleArmorAbility && a <= magmaArmorAbility
}

func (a abilityState) blocksSleep() bool {
	return a >= insomniaAbility && a <= sweetVeilAbility
}

var onSwitchAbilities = map[abilityState]func(s *slot, bs BattleState, switchIn bool){
	traceAbility:       trace,
	unnerveAbility:     unnerve,
	intimidateAbility:  intimidate,
	regeneratorAbility: regenerator,
	naturalCureAbility: naturalCure,
	drizzleAbility:     drizzle,
	droughtAbility:     drought,
	snowWarningAbility: snowWarning,
	sandStreamAbility:  sandStream,
}

func trace(s *slot, bs BattleState, switchIn bool) {
	if !switchIn {
		return
	}

	opponentMons := make([]*Pokemon, 0)
	for _, slot := range bs.getOtherSlots(s) {
		if slot.Trainer == s.Trainer {
			continue
		}
		opponentMons = append(opponentMons, slot.mon)
	}

	s.mon.Ability = opponentMons[rand.Int()%len(opponentMons)].Ability
	s.mon.trace = true
	vprintf("%s traced %s", s.mon.Base.Name, s.mon.Ability)
}

func unnerve(s *slot, bs BattleState, switchIn bool) {
	for _, otherSlot := range bs.getOtherSlots(s) {
		if s.Trainer != otherSlot.Trainer {
			otherSlot.mon.unnerved = switchIn
			otherSlot.mon.CheckItemTrigger(true, nil)
		}
	}
}

func intimidate(s *slot, bs BattleState, switchIn bool) {
	if !switchIn {
		return
	}

	for _, slot := range bs.getAllSlots() {
		if slot.Trainer == s.Trainer {
			continue
		}
		if slot.mon.Ability == innerFocusAbility {
			continue
		}
		slot.mon.changeStatStageBy(attack, -1, true)
	}
}

func regenerator(s *slot, bs BattleState, switchIn bool) {
	if switchIn || s.mon.fainted {
		return
	}

	s.mon.ChangeHpBy(s.mon.MaxHP() / 3)
}

func naturalCure(s *slot, bs BattleState, switchIn bool) {
	if switchIn {
		return
	}

	for ailment := range nonVolatileStatuses {
		delete(s.mon.Ailments, ailment)
	}
}

func drizzle(s *slot, bs BattleState, switchIn bool) {
	if !switchIn {
		return
	}
	vprintln("it started to rain")
	bs.setWeather(rainWeather)
}

func drought(s *slot, bs BattleState, switchIn bool) {
	if !switchIn {
		return
	}
	vprintln("the sunlight turned harsh")
	bs.setWeather(sunWeather)
}

func snowWarning(s *slot, bs BattleState, switchIn bool) {
	if !switchIn {
		return
	}
	vprintln("it started to hail")
	bs.setWeather(hailWeather)
}

func sandStream(s *slot, bs BattleState, switchIn bool) {
	if !switchIn {
		return
	}
	vprintln("a sandstorm brewed")
	bs.setWeather(sandstormWeather)
}

var typeConvertingAbilities = map[abilityState]func(t *pokemonType, p *int){
	aerilateAbility:    typeConvertingAbilitiesMiddleware(flyingType),
	pixilateAbility:    typeConvertingAbilitiesMiddleware(fairyType),
	galvanizeAbility:   typeConvertingAbilitiesMiddleware(electricType),
	refrigerateAbility: typeConvertingAbilitiesMiddleware(iceType),
	normalizeAbility:   normalize,
}

func normalize(t *pokemonType, p *int) {
	if *t != normalType {
		*t = normalType
		*p = *p * 6 / 5
	}
}

func typeConvertingAbilitiesMiddleware(t1 pokemonType) func(t *pokemonType, p *int) {
	return func(t2 *pokemonType, p *int) {
		if *t2 == normalType {
			*t2 = t1
			*p = *p * 6 / 5
		}
	}
}

var typeImmunityAbilities = map[abilityState]func(u *Pokemon, t pokemonType, s bool) bool{
	flashFireAbility:    flashFire,
	drySkinAbility:      drySkin,
	waterAbsorbAbility:  drySkin,
	stormDrainAbility:   stormDrain,
	voltAbsorbAbility:   voltAbsorb,
	lightningRodAbility: lightningRod,
	motorDriveAbility:   motorDrive,
	sapSipperAbility:    sapSipper,
	levitateAbility:     levitate,
}

func flashFire(p *Pokemon, t pokemonType, s bool) bool {
	if t != FireType {
		return false
	}
	p.flashFire = true
	return true
}

// still need to implement sunlight penalty
func drySkin(p *Pokemon, t pokemonType, s bool) bool {
	if t != waterType {
		return false
	}
	if s {
		return true
	}
	vprintf("%s restored health with %s", p.Base.Name, p.Ability)
	p.ChangeHpBy(p.MaxHP() / 4)
	return true
}

func stormDrain(p *Pokemon, t pokemonType, s bool) bool {
	if t != waterType {
		return false
	}
	if s {
		return true
	}
	p.changeStatStageBy(specialAttack, 1, false)
	return true
}

func voltAbsorb(p *Pokemon, t pokemonType, s bool) bool {
	if t != electricType {
		return false
	}
	if s {
		return true
	}
	vprintf("%s restored health with %s", p.Base.Name, p.Ability)
	p.ChangeHpBy(p.MaxHP() / 4)
	return true
}

func lightningRod(p *Pokemon, t pokemonType, s bool) bool {
	if t != electricType {
		return false
	}
	if s {
		return true
	}
	p.changeStatStageBy(specialAttack, 1, false)
	return true
}

func motorDrive(p *Pokemon, t pokemonType, s bool) bool {
	if t != electricType {
		return false
	}
	if s {
		return true
	}
	p.changeStatStageBy(Speed, 1, false)
	return true
}

func sapSipper(p *Pokemon, t pokemonType, s bool) bool {
	if t != grassType {
		return false
	}
	if s {
		return true
	}
	p.changeStatStageBy(attack, 1, false)
	return true
}

func levitate(p *Pokemon, t pokemonType, s bool) bool {
	return t == groundType
}

var pinchAbilities = map[abilityState]pokemonType{
	overgrowAbility: grassType,
	blazeAbility:    FireType,
	torrentAbility:  waterType,
	swarmAbility:    bugType,
}

var contactDefensiveAbilities = map[abilityState]func(u, t *slot){
	roughSkinAbility:   roughSkin,
	ironBarbsAbility:   roughSkin,
	cuteCharmAbility:   cuteCharm,
	flameBodyAbility:   flameBody,
	poisonPointAbility: poisonPoint,
	effectSporeAbility: effectSpore,
}

func roughSkin(u, t *slot) {
	change := u.mon.MaxHP() * 1 / 8
	u.mon.ChangeHpBy(-change)
	vprintf("%s was hurt by %s", u.mon.Base.Name, t.mon.Ability)
}

func cuteCharm(u, t *slot) {
	if roll(30, 100) {
		u.mon.applyAilment(infatuationAilment, nil, t)
	}
}

func flameBody(u, t *slot) {
	if roll(30, 100) {
		u.mon.applyAilment(burnAilment, nil, t)
	}
}

func poisonPoint(u, t *slot) {
	if roll(30, 100) {
		u.mon.applyAilment(poisonAilment, nil, t)
	}
}

func effectSpore(u, t *slot) {
	if u.mon.hasType(grassType) || u.mon.Ability == overcoatAbility || u.mon.Item.State == safetyGoggles {
		return
	}
	if roll(30, 100) {
		ailmentRoll := rand.Intn(30)
		if ailmentRoll <= 8 {
			u.mon.applyAilment(poisonAilment, nil, t)
		} else if ailmentRoll <= 18 {
			u.mon.applyAilment(paralysisAilment, nil, t)
		} else {
			u.mon.applyAilment(sleepAilment, nil, t)
		}
	}
}

var contactOffensiveAbilities = map[abilityState]func(u, t *slot){
	poisonTouchAbility: poisonTouch,
}

func poisonTouch(u, t *slot) {
	if roll(30, 100) {
		t.mon.applyAilment(poisonAilment, nil, u)
	}
}

func cheekPouch(mon *Pokemon) {
	if mon.Ability == cheekPouchAbility {
		restore := mon.MaxHP() / 3
		mon.ChangeHpBy(restore)
		vprintf("%s ate its cheek pouch and restored %d hp", mon.Base.Name, restore)
	}
}
