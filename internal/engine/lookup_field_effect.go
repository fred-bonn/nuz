package engine

type FieldEffect int

const (
	noneEffect FieldEffect = iota
	spikesEffect
	stealthRockEffect
	stickyWebEffect
	toxicSpikesEffect
	reflectEffect
	lightScreenEffect
	auroraVeilEffect
	tailwindEffect
	safeguardEffect
	luckyChantEffect
	gravityEffect
	trickRoomEffect
	magicRoomEffect
	wonderRoomEffect
)

var fieldEffectMap = map[string]FieldEffect{
	"spikes":       spikesEffect,
	"stealth rock": stealthRockEffect,
	"sticky web":   stickyWebEffect,
	"toxic spikes": toxicSpikesEffect,
	"reflect":      reflectEffect,
	"light screen": lightScreenEffect,
	"aurora veil":  auroraVeilEffect,
	"tailwind":     tailwindEffect,
	"safeguard":    safeguardEffect,
	"lucky chant":  luckyChantEffect,
	"gravity":      gravityEffect,
	"trick room":   trickRoomEffect,
	"magic room":   magicRoomEffect,
	"wonder room":  wonderRoomEffect,
}

func stringToFieldEffect(s string) FieldEffect {
	if e, ok := fieldEffectMap[s]; ok {
		return e
	}
	return noneEffect
}
