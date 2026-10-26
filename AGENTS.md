# AGENTS.md

## Commands

- Go 1.26+ single module (`github.com/fred-bonn/nuz`). Run commands from the repo root; the engine resolves `data/pokemon` and `data/moves` relative to the working directory.
- Use the Makefile instead of raw `go test` / `go build`:
  - `make test` — full test suite.
  - `make build` — runs the full suite first, then writes `./bin/nuz`.
  - `make run` / `make brief` — build and run the tracked demo parties (verbose / summary).
- Focused test: `make test TEST_PKG=./nuzengine TEST_ARGS="-run 'TestName'"` (escape a trailing `$` as `$$` for Make).
- `make <N>` (e.g. `make 250`) is a real target: it builds, then runs `data/player.txt data/opponent.txt --iterations N`.
- No lint or CI config exists. Format changed Go files with `gofmt`.
- Benchmarks (`go test ./nuzengine -bench ...`) run 100k-episode learning loops and need PokeAPI access/caches; they are not part of normal verification.

## Runtime data and network gotchas

- `data/` is gitignored and absent from a fresh clone. It contains the PokeAPI cache (`pokemon/`, `moves/`), `policy.json.gz`, and local demo parties.
- Cache misses are fetched from `https://pokeapi.co/api/v2` and written back, so first-time party loading needs network.
- Normal `nuzengine` tests use in-memory/dummy battle state and stay offline. CLI tests that parse full parties may populate `data/` from the API.
- `cli_test.go` reads `data/player.txt` and `data/opponent.txt`; on a fresh clone the `accepts requested data double battle` case fails until those local files exist. Do not commit the cache to fix it.
- `-f/--policy-file` overrides the CLI party arguments, weather, and battle type with values embedded in the policy. `-l` learns and writes `data/policy.json.gz` (default 100000 iterations).

## Architecture

- `main.go`: pflag CLI that calls `nuzengine.InitBattleState` / `Execute`, or `LearnMonteCarlo`.
- `nuzengine`: core engine. Single and double battles are separate implementations behind the `battleState` interface (`battle_state_single.go`, `battle_state_double.go`); feature work often needs both plus the AIs in `ai*.go`.
- Learning state keys live in `battle_state_single_key.go` and `battle_state_double_key.go`. Update them when decision-relevant state changes, or learned policies will collide.
- `internal/parser`: Showdown party lexer/parser. `internal/pokeapi`: HTTP client and API JSON mapping.
- `lookup_*.go` hold normalized metadata tables; `lookup_balancing.go` applies move overrides after API loading. `contact_moves.json` is read from the source directory at runtime.

## Conventions

- Keep identifiers unexported unless another package needs them.
- Engine identifiers are normalized lowercase strings/enums. Extend `lookup_*.go` maps and normalization (`apiName`) rather than comparing display-form strings ad hoc.
- Battle and Pokémon state is mutable. Initialize new fields in constructors and clear/restore them in `reset()` and, where relevant, `switchReset()` so iterations remain independent.
- Damage calculation is shared by move execution and AI scoring. Preserve the `roll`, `forScoring`, and `pursuit` call-site semantics, including item and ability effects.
- Tests are table-driven and should cover normal and edge cases. The demo party format is documented in `README.md`; tracked fixtures live in `showdown_demo_files/`.
