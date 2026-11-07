# AGENTS.md

## Style guide

- Keep things in one function unless composable or reusable.
- Do not extract single-use helpers preemptively. Inline the logic at the call site unless the helper is reused, hides a genuinely complex boundary, or has a clear independent name that improves the caller.
- Keep everything unexported by default. Only export when needed, and keep the number of entrypoints to packages minimal.
- Avoid using the `any` type.
- Use an empty struct when we need represent existence without storing any data (e.g. maps as hash sets).
- Before adding complexity for a speculative or vanishingly unlikely race or security edge case, explain the concrete failure mode, likelihood, and complexity cost to the user and get their buy-in. Do not silently expand scope for theoretical robustness.
- Avoid deep nested control structures as much as possible; keep the code as flat as possible.
- Reduce total variable count by inlining when a value is only used once.

```go
// Good
if err := foo(); err != nil {
    ...
}

x := foo()
if x {
    ...
}
return x

// Bad
err := foo()
if err != nil {
    return err
}
return nil
```

- Prefer short assigment statements whenever possible.

```go
// Good
x := 5

// Bad
var x = 5
```

- Avoid else statements; prefer early returns.

```go
// Good
if cond {
    return 1
}
return 0

// Bad
if cond {
    return 1
} else {
    return 0
}

```

- When a function has several validation branches or supporting details, make the main function read as the happy path and move supporting details into small helpers below it.

```go
// Good
func loadThing(input any) {
    config := requireConfig(input)
    metadata := readMetadata(input)
    return createThing(config, metadata)
}

func requireConfig(input any) {
    ...
}
```

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

## API documentation

- Document every HTTP endpoint in the `Server` section of `README.md`.
- Describe the server and its design in prose before the endpoint list. Keep front-end details out of the API section; the web page exists only to exercise the API.
- Use one bullet per endpoint in the form `- METHOD /path: summary.`
- Under each endpoint, use these bullets in order:
  - Input: the JSON body fields in a fenced `json` block, annotated as `<Type: description>`. Include required/optional status, defaults, and accepted ranges, and say `None.` when there is no body.
  - Headers: if the request has a body, set `Content-Type`; otherwise leave it empty. If there is no body and no other required headers, write `Headers: None required` on one line; otherwise use one nested bullet per required header, formatted as `Name: value`.
  - Response: the happy-path 2xx status and the returned JSON in a fenced `json` block, annotated as `<Type: description>`. For streaming endpoints, give each event type its own bullet with its JSON structure.
  - Response Headers: if the response has a body, set `Content-Type`; otherwise leave it empty. If there is no body and no other response headers, write `Response Headers: None required` on one line; otherwise use one nested bullet per response header, formatted as `Name: value`.
  - Status Codes: one nested bullet per code, with the code in backticks followed by its condition. List every status the handler can return and omit `500` and other internal failure codes.
- Document the JSON error envelope, `{"error": "..."}`.
- Include a minimal `curl` example for each endpoint or flow.
- When behaviour changes, update the affected status codes, defaults, and examples in the same change.

Example:
- POST /battles: Initialises a battle and returns its ID and the initial battle state.
  - Input: JSON with the following fields:
    ```json
    {
        player: <String: player showdown file>,
        opponent: <String: opponent showdown file>,
        double: <Boolean: default false>,
        weather: <Number: 0-4, default 0>,
    }
    ```
  - Headers:
    - `Content-Type: application/json`
  - Response: `201` with the following fields:
    ```json
    {
        id: <String: battle ID>,
        double: <Boolean: whether this is a double battle>,
        weather: <Number: weather used for the battle>,
        state: <Object: initial battle state>,
    }
    ```
  - Response Headers:
    - `X-Battle-ID`: Battle ID of the created battle.
    - `Content-Type: application/json`
  - Status Codes:
    - `201`: Battle created successfully.
    - `400`: Invalid JSON, unknown field, misformed input, or unparsable part/opponent file
    - `405`: Wrong method.
    - `503`: Concurrent battle limit reached.

## Commits

- Use conventional commit-style messages: `type(scope): summary`
- Valid types are: `feat`, `fix`, `docs`, `chore`, `refactor`, and `test`
- Scope is optional; use the affected package or area when helpful