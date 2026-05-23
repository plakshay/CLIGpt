# CLIGpt

A minimal Go CLI chatbot that streams responses from the Google Gemini API directly in your terminal.

## Features

- Interactive REPL with streaming, token-by-token responses
- Session-scoped conversation memory (user + model turns)
- Per-request timeouts via `context.WithTimeout`
- Graceful `Ctrl+C` cancellation
- `.env`-based configuration
- Single static binary, no external services beyond the Gemini API

## Project Structure

```
CLIGpt/
├── cmd/
│   ├── root.go        # Cobra root command
│   └── chat.go        # `cligpt chat` subcommand wiring
├── internal/
│   ├── ai/
│   │   └── gemini.go  # Gemini client + streaming helper
│   ├── chat/
│   │   └── session.go # REPL loop + conversation history
│   └── config/
│       └── config.go  # Env/config loading via Viper + godotenv
├── .env.example
├── go.mod
├── main.go
└── README.md
```

Each package has one responsibility:

- `cmd` — CLI surface (Cobra).
- `internal/config` — read environment, validate, return a typed `Config`.
- `internal/ai` — thin wrapper around the Gemini SDK exposing a `Stream` method.
- `internal/chat` — interactive session that owns the message history and orchestrates calls to `ai`.

## Setup

1. **Install Go 1.23+**.

2. **Get a Gemini API key** at [aistudio.google.com/apikey](https://aistudio.google.com/apikey).

3. **Configure your environment**:
   ```bash
   cp .env.example .env
   # edit .env and set GEMINI_API_KEY
   ```

4. **Download dependencies**:
   ```bash
   go mod tidy
   ```

5. **Build the binary** (optional — you can also `go run`):
   ```bash
   go build -o cligpt
   ```

## Usage

Start an interactive chat session:

```bash
./cligpt chat
```

Or without building:

```bash
go run . chat
```

Inside the session:

```
CLIGpt — type your message and press enter.
Commands: /reset to clear history, /exit to quit.

you > What is Go's zero value for a map?
ai  > It's nil. You can read from a nil map, but writing to one panics...

you > /reset
(conversation cleared)

you > /exit
bye!
```

### In-session commands

| Command          | Behavior                              |
| ---------------- | ------------------------------------- |
| `/reset`         | Clear conversation history            |
| `/exit`, `/quit` | Exit the session                      |
| `Ctrl+C`         | Cancel the current request / exit     |

## Configuration

All config is read from environment variables (loaded from `.env` if present):

| Variable                  | Default            | Description                            |
| ------------------------- | ------------------ | -------------------------------------- |
| `GEMINI_API_KEY`          | _required_         | Your Gemini API key                    |
| `GEMINI_MODEL`            | `gemini-2.0-flash` | Chat model to use                      |
| `REQUEST_TIMEOUT_SECONDS` | `60`               | Per-request timeout in seconds         |

## Architecture

The data flow is intentionally straightforward:

```
main.go
  └── cmd.Execute()              (Cobra)
        └── chatCmd.RunE
              ├── config.Load()  (Viper + godotenv)
              ├── ai.New(...)    (Gemini client wrapper)
              └── chat.New(...).Run(ctx)
                    └── loop:
                          read stdin
                          → append user message to history
                          → ai.Stream(ctx, systemPrompt, history, onChunk)
                              → prints chunks as they arrive
                          → append model message to history
```

- Conversation memory is just a `[]ai.Message` slice on the `Session` struct — no persistence, no database.
- Streaming uses the Gemini SDK's `Models.GenerateContentStream`, which exposes a Go 1.23 range-over-func iterator. Chunks are written to stdout via a callback so the chat layer stays in control of UI.
- The system prompt is passed to Gemini through `GenerateContentConfig.SystemInstruction` rather than being stored in the message history.
- Each request gets its own `context.WithTimeout`; the parent context is bound to `SIGINT`/`SIGTERM` so `Ctrl+C` cleanly aborts an in-flight stream.
- On error, the most recent user message is rolled back from history so the next attempt isn't sent twice.

## Why Cobra and Viper?

**Cobra** — the de-facto standard for Go CLIs (used by `kubectl`, `gh`, `hugo`). It gives us subcommands, `--help` output, and flag parsing for free, and it makes growing the tool (`cligpt chat`, future `cligpt config`, etc.) a one-line change. For a single-command tool we could have used `flag`, but Cobra keeps the structure tidy as commands are added.

**Viper** — handles env-var lookup, defaults, and type coercion in one place. Combined with `godotenv` for local development, it means `config.Load()` is the single source of truth for runtime configuration and the rest of the codebase just receives a typed `*Config`.

## Dependencies

- [`github.com/spf13/cobra`](https://github.com/spf13/cobra) — CLI framework
- [`github.com/spf13/viper`](https://github.com/spf13/viper) — config management
- [`github.com/joho/godotenv`](https://github.com/joho/godotenv) — `.env` loader
- [`google.golang.org/genai`](https://pkg.go.dev/google.golang.org/genai) — official Google Gen AI SDK for Go