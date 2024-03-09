package cli

type App interface {
	Init() error
	GetCommands() map[string]Command
	AddCommand(name string, cmd Command) error
	Run() error
}
