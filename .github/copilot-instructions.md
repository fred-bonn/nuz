# Nuzlocke Verifier instructions

## Testing

- Always use the Makefile rather than invoking `go test` or `go build` directly.
- Run all tests with `make test`.
- Run one test with `make test TEST_PKG=<package> TEST_ARGS="-run '<regex>'"`; escape regex end anchors as `$$` for Make.
- Every change should be testing to ensure correctness.
- When adding new tests, use table-driven cases and ensure they cover both typical and edge scenarios.
- No repository lint command is configured. Format changed Go files with `gofmt`.

## Architecture

- The repository is organized into the CLI entry point (`cmd/cli/main.go`) and HTTP server (`cmd/server`), the engine (`nuzengine`), and internal packages for parsing and API interactions.
- `nuzengine` contains the core battle logic, including state management, action resolution, and AI implementations.
- Internal packages handle specific tasks such as parsing Showdown files (`internal/parser`) and interacting with the PokéAPI (`internal/pokeapi`).
- The `data/` directory contains cached Pokémon and move metadata, as well as the learned policy for the AI.
- If metadata is missing from the `data/` caches, it will be fetched from the PokéAPI and stored locally
- Under `Usage` in the README, the flags for the CLI are specified

## Demo files

- You can add your own demo files to the `demo/` directory for testing purposes.
- The basic format for these demo files can be seen in the README.

## Repository conventions and invariants

- By default, everything should be unexported unless it needs to be accessed from outside its package.
- Engine identifiers are normalized lowercase strings or internal enum values for moves, abilities, items, types, and Pokémon names. Extend the corresponding `lookup_*.go` maps and normalization paths rather than comparing new display-form strings ad hoc.
- Battle and Pokémon state is mutable. When adding mutable per-battle or per-Pokémon fields, initialize it in constructors and restore/clear it in `reset()` and, when appropriate, `switchReset()` so iterations remain independent.
- Damage calculation is shared by real move execution and AI scoring. Preserve its `roll`, `forScoring`, and `pursuit` call-site semantics; item triggers and ability effects are part of that calculation.
