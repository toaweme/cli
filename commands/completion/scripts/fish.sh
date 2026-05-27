complete -c {{.AppName}} -f -a '({{.AppName}} __complete (commandline -cop) "" 2>/dev/null | string match -rv "^:" | string replace \\t \\t)'
