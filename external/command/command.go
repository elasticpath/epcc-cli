package command

type Command struct {
	// The keyword the command should use
	Keyword string
	// A one-line description of the command
	Description string
	// The function that will be executed
	Execute func(cmds map[string]Command, args []string) int
}
