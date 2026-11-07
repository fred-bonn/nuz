# nuz - Nuzlocke Verifier

A command-line tool and web server for simulating Pokemon battles for the purpose of developing strategies for Nuzlocke playthroughs.

## What is a Nuzlocke?

A Nuzlocke is a self-imposed ruleset for Pokemon games that turns them into roguelite games. Generally, the rules are as follows:

- For each location only the first Pokemon encounter may be caught.
- If a Pokemon faints it cannot be used anymore for the rest of the playthrough.
- If all Pokemon in the party faint the playthrough must be reset.
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
- Status, HP, and IVs are optional sections:
  - Status allows setting a non-volatile ailment, defaulting to none
  - HP allows setting the starting HP, defaulting to max HP
  - IVs allow setting each IV individually, defaulting to 31 in all stats

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

The server is a small HTTP wrapper around the battle engine for playing interactive battles over the network. Battles live in an in-memory registry keyed by a random id. The engine runs on its own goroutine and blocks whenever it needs a choice: the client starts the battle with one request, receives prompts on a long-lived newline-delimited JSON stream, and answers each prompt with a separate request. While a prompt is pending, a Monte Carlo learner runs on a copy of the battle in the background and pushes hints over the same stream. Battles are single-use: an unstarted battle expires after 2 minutes, and a started battle is aborted if its prompt goes unanswered for 2 minutes.

`make serve` runs the server in [cmd/server](cmd/server). Flags are passed through `SERVE_ARGS` (e.g. `make serve SERVE_ARGS="--addr :9000"`); the default listen address is `:8080`. Server flags are:

| Flag | Meaning |
| ---- | --- |
| `--addr <host:port>` | Listen address; default `:8080` |
| `--allow-origin <origin>` | Browser origin allowed by CORS (repeatable); CORS is disabled by default |
| `--max-battles <n>` | Maximum concurrent battles; default `100`, further creation attempts return `503` |
| `--tls-cert <file>`, `--tls-key <file>` | Serve HTTPS; both must be given together |

To serve HTTPS, pass both `--tls-cert <file>` and `--tls-key <file>` (e.g. `make serve SERVE_ARGS="--tls-cert cert.pem --tls-key key.pem"`); HSTS is then sent on every response. If a reverse proxy terminates TLS instead, it must not buffer `/battles/start` (nginx: `proxy_buffering off`) and its read timeout must exceed the 2-minute input timeout.

### API

Battles are interactive: the player is driven by the guided AI over HTTP.

- `POST /battles` initialises a battle and returns its id in the `X-Battle-ID` header and the JSON body (along with the initial `state`).
  - Input: JSON body with `player` and `opponent` (Showdown party text, required), `double` (default `false`), and `weather` (`0`–`4`, default `0`); unknown fields are rejected.
  - Headers: none required.
  - Status codes: `201` created; `400` invalid JSON, unknown field, missing party, weather out of range, or unparsable party; `405` wrong method; `503` concurrent battle limit reached.
- `POST /battles/start` starts the battle and streams its events.
  - Input: none.
  - Headers: `X-Battle-ID` required.
  - Response: `200` with `Content-Type: application/x-ndjson`; the response stays open and streams one JSON event per line:
    - `prompt`: `{"type":"prompt","text":...,"options":[...],"state":{...}}` when input is needed; `options` are the numbered choices and `state` is the full battle state.
    - `hint`: `{"type":"hint","best":N}` while a prompt is pending, sent when the learner's best option changes (omitted when nothing has been learned yet).
    - `end`: `{"type":"end","won":bool,"state":{...}}` when the battle ends.
    - `error`: `{"type":"error","error":"..."}` when the battle fails.
    - `timeout`: `{"type":"timeout","error":"..."}` when no input arrives within the 2-minute timeout.
  - Status codes: `200` streaming; `400` missing `X-Battle-ID`; `404` unknown battle; `405` wrong method; `409` battle already started.
- `POST /battles/input` answers the pending prompt.
  - Input: JSON body with `choice`, the 1-based option number.
  - Headers: `X-Battle-ID` required.
  - Status codes: `202` accepted; `400` missing `X-Battle-ID`, invalid JSON, or choice outside the current prompt; `404` unknown battle; `405` wrong method; `409` no prompt pending; `410` battle no longer running.

A choice is only accepted while its prompt is pending; the next prompt arrives on the start stream. After each answered prompt the learner moves its starting point to the new battle state and keeps what it already learned.

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
- more exhaustive coverage of field effects and interactive moves; screens and abilities that break them (e.g. Brick Break), Defog
- learning from guided examples