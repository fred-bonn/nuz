# AGENTS.md

## Style guide

- Keep things in one function unless composable or reusable.

## Testing guide

- Test every change. Run the full suite with `make test`.
- To run a focused test, use `make test TEST_PKG=<package> TEST_ARGS="-run '<regex>'"`; escape a trailing `$` as `$$` for Make.
- When adding tests, use table-driven cases covering typical and edge scenarios.
- No repository lint command is configured. Format changed Go files with `gofmt`.

## Project structure

- `.github`: GitHub configuration and Copilot instructions.
- `bin`: locally built CLI executable.
- `cmd`: application entry points.
- `cmd/cli`: command-line application.
- `cmd/server`: HTTP server.
- `data`: shared runtime data directory; PokeAPI cache and learned policies are stored here.
- `data/moves`: cached move metadata.
- `data/pokemon`: cached Pokémon metadata.
- `nuzengine`: battle engine and AI.
- `nuzengine/internal`: engine-private integrations.
- `nuzengine/internal/parser`: Showdown party parser.
- `nuzengine/internal/pokeapi`: PokéAPI client and data mapping.
- `showdown_demo_files`: tracked example parties.
- `web`: embedded web assets.
- `web/templates`: HTML templates used by the web interface.
