# nuz - Nuzlocke Verifier

A command line tool and web server for simulating Pokemon battles for the purpose of developing strategies for Nuzlocke playthroughs.

## What is a Nuzlocke?

A Nuzlocke is a self-imposed ruleset for Pokemon games that turns them into roguelite games. Generally, the rules are as follows:

- For each location only the first Pokemon encounter may be caught.
- If a Pokemon faints it cannot be used anymore for the rest of playthrough.
- If all Pokemon in the party faints the playthrough must be reset.
- Rules to prevent overleveling.

## Motivation

Losing Pokemon in a Nuzlocke playthrough can quickly compound and lead to a reset. I needed a way to more quickly evaluate how well a strategy performs and autonomously improve it.

## Quick Start

Install using the Go toolchain. This installs a binary named `cli`; the examples below call it `nuz`, which is what `make build` writes to `./bin/nuz`.

```bash
go install github.com/fred-bonn/nuz/cmd/cli@latest
```

```bash
nuz showdown_demo_files/player.txt showdown_demo_files/opponent.txt -v
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

The main entrypoint is in [cmd/cli/main.go](cmd/cli/main.go). Supported flags are:

| Flag | Meaning |
| ---- | --- |
| `-v` | Verbose logging |
| `-a <0..2>` | AI override for the player: `0=rnb`, `1=guided`, `2=random`; default `0` | 
| `-l` | Train with Monte Carlo learning and save a policy to `data/policy.json.gz` |
| `-f <path>` | Load a saved policy JSON; its embedded player and opponent parties are used |
| `-i <n>` | Number of battle iterations; defaults to `1`, or `100000` in learning mode |
| `-w <0..4>` | Weather override: `0=none`, `1=rain`, `2=sun`, `3=sandstorm`, `4=hail`; default `0` |
| `-d` | Initialise a double battle instead of the default single battle |

### Examples

Train and save a policy to `data/policy.json.gz`:

```bash
nuz -l -i 250 showdown_demo_files/player.txt showdown_demo_files/opponent.txt
```

Load a saved policy and use it statically:

```bash
nuz -f data/policy.json.gz
```

Run a battle with the random AI controlling the player:

```bash
nuz -a 2 showdown_demo_files/player.txt showdown_demo_files/opponent.txt
```

## Server

`make serve` runs the HTTP server in [cmd/server](cmd/server) (override the listen address with `make serve SERVE_ARGS="--addr :9000"`; default `:8080`). Visiting it in a browser shows a page to upload the two Showdown party files, pick single/double battle and weather, and play the battle through numbered choice buttons (templates live in [web/templates](web/templates)).

To serve HTTPS, pass both `--tls-cert <file>` and `--tls-key <file>` (e.g. `make serve SERVE_ARGS="--tls-cert cert.pem --tls-key key.pem"`); HSTS is then sent on every response. If a reverse proxy terminates TLS instead, it must not buffer `/battles/start` (nginx: `proxy_buffering off`) and its read timeout must exceed the 2 minute input timeout.

The JSON API behind the page is below. Battles are interactive: the player is driven by the guided AI over HTTP.

1. `POST /battles` accepts JSON with Showdown party text and settings, initialises a battle and replies `201` with the id in the `X-Battle-ID` header (and body). Unstarted battles expire after 2 minutes.
2. `POST /battles/start` (header `X-Battle-ID`) starts the battle. The `200` response stays open and streams newline-delimited JSON events: `{"type":"prompt","text":...,"options":[...],"state":{...}}` whenever input is needed (`options` are the numbered choices, `state` is the full battle state), then `{"type":"end","won":bool,"state":{...}}` when the battle ends (`error`/`timeout` otherwise). `201` from `POST /battles` also includes the initial `state`.
3. `POST /battles/input` (header `X-Battle-ID`, body `{"choice":N}`) answers the pending prompt with `202`; the next prompt arrives on the stream. Invalid choices are re-prompted on the stream with an error message. If no input arrives for 2 minutes the battle is aborted with a `timeout` event.
4. While a prompt is pending, a Monte Carlo learner runs in the background on a copy of the battle. Every second the stream carries `{"type":"hint","best":N}` with the number of the option it currently rates best (omitted when nothing has been learned yet), and the browser shows that choice with a green background. After each answered prompt the learner moves its starting point to the new battle state and keeps what it already learned.

```bash
curl -si -X POST localhost:8080/battles -d "$(jq -n --rawfile p showdown_demo_files/player.txt --rawfile o showdown_demo_files/opponent.txt '{player:$p,opponent:$o,double:false,weather:0}')" | grep -i x-battle-id
curl -N -X POST localhost:8080/battles/start -H "X-Battle-ID: <id>"
curl -X POST localhost:8080/battles/input -H "X-Battle-ID: <id>" -d '{"choice":1}'
```

Invalid input returns `400` with `{"error": "..."}`.

## Contributing

### Clone the repo

```bash
git clone https://github.com/fred-bonn/nuz
cd nuz
```

### Build the compiled binary

```bash
go build ./cmd/cli
```

Or (runs the tests first, writes `./bin/nuz`):

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
- more exhaustive coverage of field effects and interactive moves; screens and abilities that breaks them (e.g. Brick Break), Defog
- learning from guided examples