# mackey

**mackey** is a Go tool and library for securely storing key-value pairs in the
macOS Keychain. It replaces the need for `.env` files by keeping sensitive
configuration values (API keys, passwords, tokens, …) in the operating system's
native secure credential store.

## Features

- **Secure storage** — values are stored in the macOS Keychain, which is
  encrypted at rest and protected by the operating system.
- **Namespaced services** — group related keys under a named _service_ so
  multiple projects can share the same Keychain without conflict.
- **Shell integration** — `mackey env` prints stored key-value pairs as
  `export KEY=value` statements that can be sourced directly into a shell
  session.
- **Go library** — import the `keystore` package to load secrets at runtime in
  any Go application.

## Installation

```bash
go install github.com/dmdhrumilmistry/mackey@latest
```

Or build from source:

```bash
git clone https://github.com/dmdhrumilmistry/mackey.git
cd mackey
go build -o mackey .
```

## CLI Usage

All commands accept a `--service` / `-s` flag to specify a Keychain service
name (default: `mackey`).

```
mackey [--service <name>] <command> [args]
```

### Store a value

```bash
mackey set DATABASE_URL "postgres://user:pass@host/db"
mackey set API_KEY "sk-abc123"
```

### Retrieve a value

```bash
mackey get DATABASE_URL
# postgres://user:pass@host/db
```

### Delete a value

```bash
mackey delete DATABASE_URL
# or use the aliases: del, rm
```

### List all stored keys

```bash
mackey list
# API_KEY
# DATABASE_URL
```

### Export as shell environment variables

Source all stored keys into your current shell session:

```bash
eval "$(mackey env)"
echo $API_KEY   # sk-abc123
```

Or add it to your shell profile (`~/.zshrc`, `~/.bashrc`) to load secrets
automatically at startup:

```bash
eval "$(mackey --service myapp env)"
```

### Using a custom service namespace

```bash
mackey --service myapp set SECRET_KEY "supersecret"
mackey --service myapp get SECRET_KEY
mackey --service myapp list
```

## Go Library Usage

Import the `keystore` package to interact with the Keychain from your own Go
programs:

```go
import "github.com/dmdhrumilmistry/mackey/keystore"

func main() {
    // Store a secret.
    if err := keystore.Set("myapp", "API_KEY", "sk-abc123"); err != nil {
        log.Fatal(err)
    }

    // Retrieve a secret.
    apiKey, err := keystore.Get("myapp", "API_KEY")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(apiKey)

    // List all keys for a service.
    keys, err := keystore.List("myapp")
    if err != nil {
        log.Fatal(err)
    }
    for _, k := range keys {
        fmt.Println(k)
    }

    // Delete a key.
    if err := keystore.Delete("myapp", "API_KEY"); err != nil {
        log.Fatal(err)
    }
}
```

### Error handling

```go
val, err := keystore.Get("myapp", "MISSING_KEY")
if errors.Is(err, keystore.ErrNotFound) {
    // key does not exist
}
```

## Platform support

| Platform | Backend                                    |
|----------|--------------------------------------------|
| macOS    | macOS Keychain (via Security framework)    |
| Linux    | Secret Service API (e.g. GNOME Keyring)    |
| Windows  | Windows Credential Manager                 |

> **Note:** Although the primary target is macOS, the underlying
> [`go-keyring`](https://github.com/zalando/go-keyring) library supports Linux
> and Windows as well.

## License

MIT
