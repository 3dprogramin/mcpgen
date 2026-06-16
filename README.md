# mcpgen

A tiny CLI that generates project-local `.mcp.json` files from a bundled catalog
of [MCP](https://modelcontextprotocol.io) servers — so you stop copy-pasting and
reformatting server configs every time you start a new project.

## Install

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

Catalog entries ship with **placeholders** for secrets and paths (API keys,
connection strings, vault paths). After generating, open `.mcp.json` and replace
any `YOUR_*` / `/path/to/...` values with your real ones.

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

### Contributing a server

1. Add an entry to `servers.json` (use placeholders, never real secrets).
2. Run `go build -o mcpgen .` and verify with `mcpgen list`.
3. Open a pull request.

## License

[MIT](./LICENSE)
