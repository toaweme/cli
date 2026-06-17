Output demo.gif
Set FontSize 16
Set Width 1000
Set Height 600
Set Padding 20
Set Theme Dracula

Hide
  Type "unset PROMPT_COMMAND"
  Enter
  Type "cd /tmp/cli-demo"
  Enter
  Type "clear"
  Enter
Show

Sleep 500ms
Type "# toaweme/cli - Effortless Go CLI apps"
Enter
Type "# Run these yourself from examples/ dir"
Sleep 1s
Enter
Sleep 1s

# --- Basic: help & info ---
Enter
Type "# 1. Basic app with built-in help"
Enter
Sleep 1s

Type "./bin/basic --help"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

Enter
Type "./bin/basic info"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

# --- Greet: positional args, flags, env ---
Type "# 2. Positional args, flags, and env binding"
Enter
Sleep 1s

Type "./bin/greet greet Ada"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

Enter
Type "./bin/greet greet Ada --shout"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

Enter
Type "./bin/greet greet Ada -s -r 3"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

Enter
Type "GREET_NAME=World ./bin/greet greet"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

Enter
Type "./bin/greet greet --help"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

# --- Deploy: subcommand trees ---
Enter
Type "# 3. Subcommand trees"
Enter
Sleep 1s

Type "./bin/deploy deploy"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

Enter
Type "./bin/deploy deploy staging v1.2.3"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

Enter
Type "./bin/deploy deploy production v1.2.3 --dry-run"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

# --- Full: build, serve, db, config ---
Type "# 4. Full-featured app: build, serve, subcommands"
Enter
Sleep 1s

Type "./bin/full --help"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

Enter
Type "./bin/full build"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

Enter
Type "./bin/full build -o ./dist --tags=integration,e2e --verbose"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

Enter
Type "./bin/full serve --help"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

Enter
Type "./bin/full serve --help --help-format=pretty"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

Enter
Type "./bin/full serve --help --help-format=md"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

# --- Secret masking ---
Enter
Type "# 5. Secret masking with --help-values"
Enter
Sleep 2s

Type "SERVER_TOKEN=sk-abc123secret ./bin/server start --help-values"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

# --- DB subcommands ---
Enter
Type "# 6. Nested subcommands"
Enter
Sleep 1s

Type "./bin/full db"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

Enter
Type "./bin/full db migrate -n 3"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

Enter
Type "./bin/full db migrate --dry-run"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

# --- Version ---
Enter
Type "# 7. Built-in version flag"
Enter
Sleep 1s

Type "./bin/full --version"
Sleep 500ms
Enter
Sleep 3s

Hide
  Type "clear"
  Enter
Show

# --- Outro ---
Enter
Type "# github.com/toaweme/cli"
Sleep 500ms
Enter
Sleep 3s
