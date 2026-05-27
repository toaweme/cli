package help

import (
	"testing"

	"github.com/toaweme/cli"
)

type stubCommand struct {
	cli.BaseCommand[any]
	help string
}

var _ cli.Command[any] = (*stubCommand)(nil)

func (s *stubCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error { return nil }
func (s *stubCommand) Help() string                                  { return s.help }

func newStub(name, help string, subs ...cli.Command[any]) cli.Command[any] {
	cmd := &stubCommand{help: help}
	cmd.Name(name)
	for _, sub := range subs {
		cmd.Add(sub.Name(""), sub)
	}
	return cmd
}

func newStubNamed(name, help string) cli.Command[any] {
	cmd := &stubCommand{help: help}
	cmd.Name(name)
	return cmd
}

func Test_FilterCommands(t *testing.T) {
	migrate := newStubNamed("migrate", "Run migrations")
	seed := newStubNamed("seed", "Seed database")
	reset := newStubNamed("reset", "Reset database")
	db := newStub("db", "", migrate, seed, reset)

	show := newStubNamed("show", "Show config")
	set := newStubNamed("set", "Set config")
	config := newStub("config", "", show, set)

	build := newStub("build", "Build the project")
	serve := newStub("serve", "Start server")

	all := []cli.Command[any]{build, serve, config, db}

	tests := []struct {
		name       string
		filters    []string
		wantNames  []string
		wantCount  int
		checkSubs  map[string]int
	}{
		{
			name:      "top-level command",
			filters:   []string{"build"},
			wantNames: []string{"build"},
			wantCount: 1,
		},
		{
			name:      "multiple top-level commands",
			filters:   []string{"build", "serve"},
			wantNames: []string{"build", "serve"},
			wantCount: 2,
		},
		{
			name:      "parent by name returns with all subs",
			filters:   []string{"db"},
			wantNames: []string{"db"},
			wantCount: 1,
			checkSubs: map[string]int{"db": 3},
		},
		{
			name:      "subcommand path filters parent to matched sub",
			filters:   []string{"db migrate"},
			wantNames: []string{"db"},
			wantCount: 1,
			checkSubs: map[string]int{"db": 1},
		},
		{
			name:      "multiple subcommand paths",
			filters:   []string{"db migrate", "db seed"},
			wantNames: []string{"db"},
			wantCount: 1,
			checkSubs: map[string]int{"db": 2},
		},
		{
			name:      "nonexistent command returns empty",
			filters:   []string{"deploy"},
			wantCount: 0,
		},
		{
			name:      "mixed top-level and subcommand path",
			filters:   []string{"build", "config show"},
			wantNames: []string{"build", "config"},
			wantCount: 2,
			checkSubs: map[string]int{"config": 1},
		},
		{
			name:      "whitespace in filter is trimmed",
			filters:   []string{" build "},
			wantNames: []string{"build"},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterCommands(all, tt.filters)

			if len(result) != tt.wantCount {
				t.Fatalf("expected %d commands, got %d", tt.wantCount, len(result))
			}

			for i, wantName := range tt.wantNames {
				if i >= len(result) {
					break
				}
				got := result[i].Name("")
				if got != wantName {
					t.Fatalf("result[%d]: expected name %q, got %q", i, wantName, got)
				}
			}

			for name, wantSubCount := range tt.checkSubs {
				for _, cmd := range result {
					if cmd.Name("") == name {
						subs := cmd.Commands()
						if len(subs) != wantSubCount {
							t.Fatalf("command %q: expected %d subs, got %d", name, wantSubCount, len(subs))
						}
					}
				}
			}
		})
	}
}

func Test_FilterCommands_DoesNotMutateOriginal(t *testing.T) {
	migrate := newStubNamed("migrate", "Run migrations")
	seed := newStubNamed("seed", "Seed database")
	db := newStub("db", "", migrate, seed)

	all := []cli.Command[any]{db}

	_ = FilterCommands(all, []string{"db migrate"})

	if len(all[0].Commands()) != 2 {
		t.Fatalf("original command was mutated: expected 2 subs, got %d", len(all[0].Commands()))
	}
}
