# Nuzlocke Verifier

This project simulates Pokémon battles, learns a per-state action policy for the player side, and can save that policy to disk for later replay.

## What has been added

### AI opponents

- `rnbAi` ("Run & Bun"), an AI based on the Run & Run romhack ([struct_ai_rnb.go](struct_ai_rnb.go))
- `randomAi`, a baseline random action chooser ([struct_ai_random.go](struct_ai_random.go))
- `guidedAi`, an interactive AI that prompts the user ([struct_ai_guided.go](struct_ai_guided.go))
- `learningAi`, reinforcement learning for the player trainer using discretized battle states ([struct_ai_learning.go](struct_ai_learning.go))
- a static policy executor used when loading a saved policy

### Data and persistence

- a Showdown-style party file parser ([internal/parser](internal/parser/parser.go)) with its own lexer and validation
- a PokeAPI client for fetching Pokémon and move data, with local JSON caching under [data/pokemon](data/pokemon) and [data/moves](data/moves) ([internal/pokeapi](internal/pokeapi/client.go))
- persisted policy storage under a `policies/` directory
- command-line support for training, saving, and reusing learned policies

## CLI usage 

The main entrypoint is in [main.go](main.go). Supported flags are:

| Flag | Meaning |
| --- | --- |
| `-p`, `--player-learning-ai` | Train the player using the learning AI while the opponent keeps the RNB AI logic |
| `-g`, `--player-guided-ai` | Prompt for the player's action each turn (implies `--verbose`) |
| `-f`, `--policy-file <path>` | Load a saved policy JSON and use it as a static policy for the player; uses the player/opponent parties embedded in the policy, so `<player_showdown> <opponent_showdown>` must be omitted |
| `-s`, `--save-policy` | Save the learned policy to `policies/` using the two input file names |
| `-i`, `--iterations <n>` | Number of battle repetitions to run for training or statistics |
| `-w`, `--weather <0..4>` | Weather override: `0=none`, `1=rain`, `2=sun`, `3=sandstorm`, `4=hail` |
| `-v`, `--verbose` | Verbose logging |

Examples:

```bash
go run . -p -i 250 data/player.txt data/rnb_trainer_1.txt
```

Train and save a policy:

```bash
go run . -p -s -i 250 data/player.txt data/rnb_trainer_1.txt
```

Load a saved policy and use it statically. The player and opponent parties come from the policy file itself, so no input files are given:

```bash
go run . -f policies/player__vs__rnb_trainer_1.json -i 1
```

The long-form flags remain supported too, so these are equivalent:

```bash
go run . --player-learning-ai --iterations 250 data/player.txt data/rnb_trainer_1.txt
```

The policy file is named using the input filenames:

```text
policies/<player_name>__vs__<opponent_name>.json
```

## Saved policy format

Policies are stored as JSON and live under the `policies/` directory. The files include:

- `player_party`: the entire contents of the player's Showdown party file (i.e. `cat data/player.txt`)
- `opponent_party`: the entire contents of the opponent's Showdown party file
- `policy`: the state-to-action map
- `scores`: learned action scores per state
- `counts`: action usage counts per state
- `created_at`: save timestamp
- `version`: policy version
- `metadata`: source input file names

`player_party`/`opponent_party` hold the raw, unparsed Showdown text (items, IVs, moves, everything), so a saved policy is self-contained: `--policy-file`/`-f` runs that text back through the same lexer/parser used for `<player_showdown>` files.

Example shape:

```json
{
  "player_party": "Horsea @ Oran Berry\nLevel: 17\nModest Nature\nAbility: Swift Swim\n- Bubble Beam\n- Twister\n",
  "opponent_party": "Dwebble @ Salac Berry\nLevel: 17\nAdamant Nature\nAbility: Sturdy\n- Knock Off\n- Sticky Web\n",
  "policy": {
    "{\"player_pokemon\":\"horsea\",...}": [
      "move:bubble beam",
      "switch:staravia"
    ]
  },
  "scores": {
    "{\"player_pokemon\":\"horsea\",...}": {
      "move:bubble beam": 142.5,
      "switch:staravia": -45.2
    }
  },
  "counts": {
    "{\"player_pokemon\":\"horsea\",...}": {
      "move:bubble beam": 17,
      "switch:staravia": 4
    }
  }
}
```

## Input party format

The parser accepts Showdown-style party files, like the examples in [data/player.txt](data/player.txt) and [data/rnb_trainer_1.txt](data/rnb_trainer_1.txt).

### Basic format

```text
Pokemon Name @ Item
Level: N
Nature Nature
Ability: Ability
Status: Status
HP: N
IVs: 31 Atk / 31 Def / 31 SpA / 31 SpD / 31 Spe
- Move 1
- Move 2
- Move 3
- Move 4
```

### Example

```text
Horsea @ Oran Berry
Level: 17
Modest Nature
Ability: Swift Swim
- Bubble Beam
- Twister
```

### Notes

- `@ Item` is optional
- `Level`, `Nature`, and `Ability` are required
- status, HP, and IVs are optional sections:
  - status allows the set a non-volatile ailment, default to none
  - HP allows to set the starting HP for the, default to max HP
  - IVs allows for individually setting each IV, default to 31 in all stats
- moves are listed as `- Move Name`
- the parser normalizes names, strips punctuation, and reads the move list into the battle state

## Behavior of loaded policies

When a policy is loaded via `--policy-file`/`-f`, it is used as a static policy, not as a live learning AI. That means:

- it does not accumulate new battle outcomes during evaluation
- it does not decay or reset the learned scores while the battle runs
- it uses the saved score table to select the best action for the current state
- it remains deterministic for a given input file and policy file

This is specifically important when comparing a single-battle loaded run to a long training run. A single loaded battle with a saved policy is expected to behave consistently, while the training AI continues to improve over many iterations.

## TODO

- more exhaustive coverage of abilities and items; several common competitive abilities/items are missing (e.g. Multiscale, Protean, priority-negating abilities)
- support for double battles; only single 1v1 slots per side
- more exhaustive coverage of field effects and interactive moves; screens and abilities that breaks them (e.g. Brick Break), Defog