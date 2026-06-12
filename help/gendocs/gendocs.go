// Package gendocs renders an application's command tree to documentation files on disk,
// in every help format the app supports. It drives the same in-process renderers as the
// --help-format flag (no subprocess), so the generated docs are a byte-for-byte match
// for what users see, and stay in sync with the command structs they are generated from.
package gendocs

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/help"
)

// Options controls a documentation generation run.
type Options struct {
	// AppName is the binary name shown in usage lines and example invocations.
	AppName string
	// Commands is the app's command tree (typically App.Commands()).
	Commands []cli.Command[any]
	// Codecs are the output codecs registered on the app (typically App.OutputFormats()),
	// so docs are also emitted in every custom format (yaml, toml, ...).
	Codecs []cli.OutputCodec
	// Dir is the output directory. Files are written under Dir/AppName.
	Dir string
	// PerCommand also writes one file per command (and subcommand) under Dir/AppName/commands,
	// in the human-readable and json formats.
	PerCommand bool
}

// Generate renders the command tree to files under opts.Dir/opts.AppName and returns the
// paths written, relative to opts.Dir. The whole tree is emitted once per format
// (markdown, plain text, json, json schema, and every registered codec);
// with PerCommand, each command is additionally emitted on its own.
func Generate(opts Options) ([]string, error) {
	if opts.AppName == "" {
		return nil, errors.New("failed to generate docs: app name is required")
	}

	base := filepath.Join(opts.Dir, opts.AppName)
	formatNames := codecFormatNames(opts.Codecs)

	var written []string

	treeFiles, err := renderTree(opts.AppName, opts.Commands, opts.Codecs, formatNames, base)
	if err != nil {
		return written, err
	}
	written = append(written, treeFiles...)

	if opts.PerCommand {
		for _, path := range commandPaths(opts.Commands, nil) {
			cmdDir := filepath.Join(append([]string{base, "commands"}, path...)...)
			filtered := help.FilterCommands(opts.Commands, path)
			files, err := renderCommand(opts.AppName, filtered, formatNames, cmdDir)
			if err != nil {
				return written, err
			}
			written = append(written, files...)
		}
	}

	rel := make([]string, len(written))
	for i, p := range written {
		r, relErr := filepath.Rel(opts.Dir, p)
		if relErr != nil {
			r = p
		}
		rel[i] = r
	}
	return rel, nil
}

// renderTree writes one documentation file per format for the whole command tree.
func renderTree(appName string, commands []cli.Command[any], codecs []cli.OutputCodec, formatNames []string, dir string) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create docs directory %q: %w", dir, err)
	}

	var written []string

	textFiles := map[string]string{
		"help.md":   "md",
		"help.txt":  "plain",
		"help.json": "json",
	}
	for name, format := range textFiles {
		path := filepath.Join(dir, name)
		if err := writeFile(path, renderFormat(appName, commands, format, formatNames)); err != nil {
			return written, err
		}
		written = append(written, path)
	}

	schemaPath := filepath.Join(dir, "schema.json")
	if err := writeFile(schemaPath, renderJSONSchema(commands)); err != nil {
		return written, err
	}
	written = append(written, schemaPath)

	for _, codec := range codecs {
		names := cli.FormatAliases(codec)
		if len(names) == 0 {
			continue
		}
		data, err := renderCodec(commands, codec)
		if err != nil {
			return written, err
		}
		path := filepath.Join(dir, "help."+names[0])
		if err := writeFile(path, data); err != nil {
			return written, err
		}
		written = append(written, path)
	}

	return written, nil
}

// renderCommand writes the human-readable and json docs for a single (filtered) command.
func renderCommand(appName string, commands []cli.Command[any], formatNames []string, dir string) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create docs directory %q: %w", dir, err)
	}

	var written []string
	files := map[string]string{
		"help.md":   "md",
		"help.json": "json",
	}
	for name, format := range files {
		path := filepath.Join(dir, name)
		if err := writeFile(path, renderFormat(appName, commands, format, formatNames)); err != nil {
			return written, err
		}
		written = append(written, path)
	}
	return written, nil
}

func renderFormat(appName string, commands []cli.Command[any], format string, formatNames []string) []byte {
	var b bytes.Buffer
	switch format {
	case "json":
		help.DisplayHelpJSON(&b, commands)
	default:
		help.DisplayHelpAgent(&b, help.AgentOptions{
			AppName:  appName,
			Format:   format,
			Commands: commands,
			Formats:  formatNames,
		})
	}
	return b.Bytes()
}

func renderJSONSchema(commands []cli.Command[any]) []byte {
	var b bytes.Buffer
	help.DisplayHelpJSONSchema(&b, commands)
	return b.Bytes()
}

func renderCodec(commands []cli.Command[any], codec cli.OutputCodec) ([]byte, error) {
	var b bytes.Buffer
	if err := help.DisplayHelpEncoded(&b, commands, codec); err != nil {
		return nil, fmt.Errorf("failed to render docs as %q: %w", codec.Extension(), err)
	}
	return b.Bytes(), nil
}

// codecFormatNames returns the primary --help-format name of each codec, in registration
// order, for the --help-format hint shown in the generated docs.
func codecFormatNames(codecs []cli.OutputCodec) []string {
	var names []string
	seen := make(map[string]bool, len(codecs))
	for _, codec := range codecs {
		aliases := cli.FormatAliases(codec)
		if len(aliases) == 0 || seen[aliases[0]] {
			continue
		}
		seen[aliases[0]] = true
		names = append(names, aliases[0])
	}
	return names
}

// commandPaths walks the command tree and returns the path of every command and subcommand
// (e.g. ["build"], ["db", "migrate"]), depth-first.
func commandPaths(commands []cli.Command[any], prefix []string) [][]string {
	var paths [][]string
	for _, cmd := range commands {
		path := append(append([]string{}, prefix...), cmd.Name(""))
		paths = append(paths, path)
		paths = append(paths, commandPaths(cmd.Commands(), path)...)
	}
	return paths
}

func writeFile(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write docs file %q: %w", path, err)
	}
	return nil
}
