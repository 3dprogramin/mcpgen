# Render with: make demo   (needs https://github.com/charmbracelet/vhs)
Output "_img/demo.gif"

Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 800
Set Padding 20
Set Theme "Dracula"
Set TypingSpeed 60ms

# Put the freshly built binary on PATH and start clean (hidden from the GIF).
Hide
Type `export PATH="$PWD:$PATH"; rm -f .mcp.json; clear`
Enter
Show

Type "mcpgen generate"
Sleep 500ms
Enter
Sleep 1500ms

# Filter to chrome and select it.
Type "chrome"
Sleep 800ms
Space
Sleep 700ms

# Clear the filter, search for mongo, select it.
Backspace 6
Sleep 500ms
Type "mongo"
Sleep 800ms
Space
Sleep 700ms

Enter
Sleep 1s

# Override the Chrome port.
Type "-y chrome-devtools-mcp@latest --browser-url=http://127.0.0.1:9333"
Sleep 400ms
Enter
Sleep 800ms

# Keep MongoDB's defaults.
Enter
Sleep 1500ms

Type "cat .mcp.json"
Enter
Sleep 3s

# Clean up (hidden).
Hide
Type "rm -f .mcp.json"
Enter
