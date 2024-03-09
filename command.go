package cli

type Command interface {
	Help() any
	Run(vars map[string]any) error
	Validate(vars map[string]any) error
	Structure() any
}
