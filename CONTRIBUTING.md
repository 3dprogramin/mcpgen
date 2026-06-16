# Contributing to mcpgen

Thanks for helping grow the catalog! The most valuable contribution is usually
**adding a new MCP server** so others can generate it with one command.

## Add a server to the catalog

1. Open [`pkg/catalog/servers.json`](./pkg/catalog/servers.json) and add an
   entry under `servers`:

   ```json
   "my-server": {
     "description": "Short one-line summary shown in `mcpgen list`",
     "config": {
       "type": "stdio",
       "command": "npx",
       "args": ["-y", "my-mcp-server", "--token=YOUR_TOKEN"],
       "env": {}
     }
   }
   ```

   For an SSE server, use the `url` form instead:

   ```json
   "my-server": {
     "description": "...",
     "config": { "type": "sse", "url": "http://127.0.0.1:1234" }
   }
   ```

2. **Use placeholders, never real secrets.** Replace API keys, tokens, paths and
   connection strings with obvious placeholders like `YOUR_TOKEN`,
   `/path/to/your/data`, `mongodb://127.0.0.1:27017/YOUR_DATABASE`. CI rejects
   anything that looks like a real credential.

3. Verify locally:

   ```sh
   make check          # gofmt, vet, tests, build
   go run . list       # confirm your server shows up
   go run . generate my-server
   ```

4. Open a pull request.

## Code changes

- Run `make check` before pushing - CI runs the same steps and will fail on
  unformatted code, vet errors, or failing tests.
- Add or update tests for behavior changes.

## Commit messages

Use single-line [Conventional Commits](https://www.conventionalcommits.org):
`feat: ...`, `fix: ...`, `docs: ...`, `test: ...`, `chore: ...`.
