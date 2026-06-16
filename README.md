# mcpgen

[![CI](https://github.com/3dprogramin/mcpgen/actions/workflows/ci.yml/badge.svg)](https://github.com/3dprogramin/mcpgen/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/3dprogramin/mcpgen)](https://github.com/3dprogramin/mcpgen/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

A tiny CLI that generates project-local `.mcp.json` files from a bundled catalog
of [MCP](https://modelcontextprotocol.io) servers — so you stop copy-pasting and
reformatting server configs every time you start a new project.

![demo](./demo.gif)

## Install

Download a prebuilt binary for your platform from the
[latest release](https://github.com/3dprogramin/mcpgen/releases/latest), or:

```sh
go install github.com/3dprogramin/mcpgen@latest
```

Or build from source:

```sh
git clone https://github.com/3dprogramin/mcpgen
cd mcpgen
go build -o mcpgen .
```

## Usage

List the servers available in the catalog:

```sh
mcpgen list
```

Generate (or extend) `./.mcp.json` with one or more servers:

```sh
mcpgen generate chrome-devtools mongodb
```

- If `./.mcp.json` doesn't exist, it's created.
- If it exists, the selected servers are **merged in**.
- If a server is already defined, mcpgen refuses to clobber it — pass `--force`
  (`-f`) to overwrite:

```sh
mcpgen generate chrome-devtools --force
```

### Interactive mode

Run `generate` with no server names to pick servers and then set any per-server
args:

```sh
mcpgen generate
```

On a terminal you get an arrow-key checkbox list:

```
Select MCP servers — ↑/↓ move · space toggle · a all · enter confirm · q quit
> [x] burp             Burp Suite — SSE bridge to Burp's MCP endpoint
  [ ] chrome-devtools  Chrome DevTools — drive and inspect a Chrome instance
  [x] mongodb          MongoDB — query and manage a MongoDB database
  ...
```

When stdin/stdout isn't a terminal (pipes, CI), it falls back to a numbered
prompt: `Select servers (e.g. 1 3, 1-3, or 'all')`.

After selecting, each server with args shows its current args and lets you type
a replacement. The current args are shown so you can copy and tweak them:

```
Args for "chrome-devtools"
  current: -y chrome-devtools-mcp@latest --browser-url=http://127.0.0.1:9222
  new args (replaces all, blank to keep):
```

Whatever you type **replaces the whole arg list**; leave it blank to keep the
defaults. (This differs from the command-line form below, which merges.)

### Overriding args

Arg overrides apply to **exactly one server at a time**. If you pass any args,
you must name a single server — naming two or more with args is an error. To
customize several servers, use interactive mode or run `generate` once per
server.

Any extra args after the single server name are merged into its `args`:

```sh
mcpgen generate chrome-devtools --browser-url=http://127.0.0.1:9333
```

- A `--flag=value` override **replaces** a matching flag already in the config —
  whether it's stored joined (`--flag=old`) or split (`--flag`, `old`).
- Anything else is **appended** (e.g. `--headless`).

```sh
mcpgen generate mongodb --connectionString=mongodb://127.0.0.1:27017/mydb
```

#### Reserved flags and the `--` separator

`-f` / `--force` (and `--` itself) are reserved by mcpgen, so before the
separator they control mcpgen, not the server. Everything **after `--`** is
passed to the server verbatim. Use it to pass args that start without a dash, or
args that would otherwise collide with a reserved flag:

```sh
# value doesn't start with "-"
mcpgen generate mongodb -- mongodb://127.0.0.1:27017/mydb

# pass a literal --force to the server's command
mcpgen generate chrome-devtools -- --force
```

Catalog entries ship with **placeholders** for secrets and paths (API keys,
connection strings, vault paths). After generating, open `.mcp.json` and replace
any `YOUR_*` / `/path/to/...` values with your real ones.

## Output & colors

On an interactive terminal mcpgen shows a banner and uses color. When stdout
isn't a terminal (pipes, redirects, CI) or `NO_COLOR` is set, output is plain
text with no banner or escape codes.

## Catalog

The catalog lives in [`servers.json`](./servers.json) and is compiled into the
binary via `go:embed`. Each entry is just a name, a description, and the raw MCP
server config:

```json
{
  "servers": {
    "burp": {
      "description": "Burp Suite — SSE bridge to Burp's MCP endpoint",
      "config": {
        "type": "sse",
        "url": "http://127.0.0.1:9876"
      }
    }
  }
}
```

Both `stdio` servers (`command` + `args` + `env`) and `sse` servers (`url`) are
supported — the config is passed through verbatim, so any valid MCP server shape
works.

## Contributing

The catalog is meant to be community-maintained — adding a server is the easiest
and most useful contribution. See [CONTRIBUTING.md](./CONTRIBUTING.md) for the
entry template and rules (always use placeholders, never real secrets).

## Development

```sh
make check      # gofmt check, go vet, tests (-race), build
make test       # tests only
make build      # build ./mcpgen
make demo       # render demo.gif (needs https://github.com/charmbracelet/vhs)
make snapshot   # local GoReleaser build, no publish (needs goreleaser)
```

Releases are cut by pushing a `vX.Y.Z` tag; a GitHub Actions workflow runs
GoReleaser to build cross-platform binaries and publish a GitHub Release.

## License

[MIT](./LICENSE)
