# Nuzlocke Verifier instructions

## Toolchain and commands

- This is a Go module (`github.com/fred-bonn/nuz`) targeting Go 1.26.1. The only external dependency is `spf13/pflag`.
- Always use the Makefile rather than invoking `go test` or `go build` directly. Run all tests with `make test`.
- Run one test with `make test TEST_PKG=<package> TEST_ARGS="-run '<regex>'"`; escape regex end anchors as `$$` for Make:
  - Root CLI test: `make test TEST_PKG=. TEST_ARGS="-run '^TestRunReturnsTheExpectedExitCodeForCLIArguments$$'"`
  - Engine test/subtest: `make test TEST_PKG=./nuzengine TEST_ARGS="-run '^TestCalculateDamageMatchesGameRollRange$$/^bubble_beam$$'"`
  - Internal package tests use their package path, for example `make test TEST_PKG=./nuzengine/internal/parser TEST_ARGS="-run '^TestReadShowdownFileCases$$'"`.
- `make build` first runs the full test suite and writes `bin/nuz`. `make run` executes the demo Showdown parties with verbose logging.
- No repository lint command is configured. Format changed Go files with `gofmt`.

## Architecture

- `main.go` owns flag parsing and reads the two Showdown party files. It delegates battle setup to `nuzengine.InitBattleState` and execution/statistics to `nuzengine.Execute`.
- `nuzengine/config.go` is the input boundary: `internal/parser` parses Showdown text, then the config loads Pokémon and move metadata, builds engine `pokemon` values, and validates abilities, items, stats, and levels.
- Metadata is read from relative `data/pokemon/` and `data/moves/` caches. Missing entries are fetched from PokéAPI through `internal/pokeapi` and written to those caches. These generated root-level `data/` directories are ignored, so commands run from the repository root and from the `nuzengine` package can use different cache locations.
- The `battleState` interface abstracts single and double battles. Its implementations own active slots, trainers, field effects, weather, the action queue, and snapshots used by `reset()` for repeated simulations. A turn gathers actions, sorts and executes them, then resolves end-of-turn effects and queued replacement actions.
- Actions are concrete `moveAction`, `switchAction`, and `replaceAction` values. They execute against `battleState`; priority and effective speed determine queue order. Move resolution, residual effects, switching, and fainting can enqueue additional actions, so preserve this queue-based flow when adding battle mechanics.
- AI implementations share the `ai` interface. Run & Bun is the default heuristic, guided AI is interactive, and Monte Carlo learning records state/action trajectories into a Q-map. Learning writes the compressed policy to `data/policy.json.gz`; loading a policy also supplies its embedded parties and weather.

## Repository conventions and invariants

- Most engine types and helpers are deliberately unexported. Tests in `nuzengine` use the same package so they can construct minimal battle state directly; keep new engine tests there unless testing the CLI boundary. Write Go unit tests using table-driven cases.
- Engine identifiers for moves, abilities, items, types, and Pokémon names are normalized lowercase strings or internal enum values. Extend the corresponding `lookup_*.go` maps and normalization paths rather than comparing new display-form strings ad hoc.
- Battle and Pokémon state is mutable. When adding mutable per-battle or per-Pokémon fields, initialize it in constructors and restore/clear it in `reset()` and, when appropriate, `switchReset()` so iterations remain independent.
- Damage calculation is shared by real move execution and AI scoring. Preserve its `roll`, `forScoring`, and `pursuit` call-site semantics; item triggers and ability effects are part of that calculation.
- Keep Showdown parsing compatible with its ordered format: name/item, level, nature, and ability are required; status, HP, IVs, and up to four move lines are optional. Parser errors are line-aware and should remain actionable.
