# nuz - Nuzlocke Verifier

A command line tool for simulating Pokemon battles for the purpose of developing strategies for Nuzlocke playthroughs.

## What is a Nuzlocke?

A Nuzlocke is a self-imposed ruleset for Pokemon games that turns them into roguelite games. Generally, the rules are as follows:

- For each location only the first Pokemon encounter may be caught.
- If a Pokemon faints it cannot be used anymore for the rest of playthrough.
- If all Pokemon in the party faints the playthrough must be reset.
- Rules to prevent overleveling.

## Motivation

Losing Pokemon in a Nuzlocke playthrough can quickly compound and lead to a reset. I needed a way to more quickly evaluate how well a strategy performs and autonomously improve it.

## Quick Start

Install nuz using the Go toolchain.

```bash
go install github.com/fred-bonn/nuz

```

```bash
nuz showdown_demo_files/player.text showdown_demo_files/opponent.txt -v
```

### Input party format

There is a parser included in nuz that accepts Showdown-style party files, e.g. `player.txt`, `opponent.txt` from the example.

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

## Usage

The main entrypoint is in [main.go](main.go). Supported flags are:

| Flag | Meaning |
| ---- | --- |
| `-v` | Verbose logging |
| `-a <0..3>` | AI override for the player: `0=rnb`, `1=learning`, `2=guided`, `3=random`; default `0` | 
| `-f <path>` | Load a saved policy JSON and use it as a static policy for the player |
| `-i <n>` | Number of battle repetitions; default `1` |
| `-w <0..4>` | Weather override: `0=none`, `1=rain`, `2=sun`, `3=sandstorm`, `4=hail`; default `0` |

Examples:

Train and save a policy:

```bash
nuz -a 1 showdown_demo_files/player.text showdown_demo_files/opponent.txt
```

Load a saved policy and use it statically:

```bash
nuz -f policies/player__vs__opponent.json
```

Simulate the battle as REPL, prompting the player's action each turn:

```bash
nuz -a 2 showdown_demo_files/player.text showdown_demo_files/opponent.txt
```

## Contributing

### Clone the repo

```bash
git clone https://github.com/fred-bonn/nuz
cd nuz
```

### Build the compiled binary

```bash
go build
```

Or:

```bash
make build
```

### Run the test suite

```bash
go test ./...
```

Or:

```bash
make test
```

### Submit a pull request

If you'd like to contribute, please fork the repository and open a pull request to the `main` branch.

## TODO

- more exhaustive coverage of abilities and items; several common competitive abilities/items are missing (e.g. Multiscale, Protean, priority-negating abilities)
- support for double battles; only single 1v1 slots per side
- more exhaustive coverage of field effects and interactive moves; screens and abilities that breaks them (e.g. Brick Break), Defog
- separate the battle engine from the front-end so it can be re-used for other purposes
- improving the policy creation; better/less coarse discrete states
- learning from guided examples