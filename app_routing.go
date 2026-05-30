package cli

func (c *app) matchCommandByArgs(args []string) (Command[any], []string, []string, error) {
	var command Command[any]
	var commandNameIndexes []int

	for a := 0; a < len(args); a++ {
		// previous arg is a command
		// assert if this arg is a sub command
		if command != nil {
			subCommand := c.matchCommandByName(args[a], command.Commands())
			if subCommand != nil {
				command = subCommand
				commandNameIndexes = append(commandNameIndexes, a)
				continue
			}

			// no sub command found, but
			// we have a command already so let's keep it and try further args
			break
		}

		cmd := c.matchCommandByName(args[a], c.commands)
		if cmd != nil {
			command = cmd
			commandNameIndexes = append(commandNameIndexes, a)
			continue
		}
	}

	if command == nil {
		return nil, nil, nil, ErrCommandNotFound
	}

	// create a new slice that excludes the command args
	// we don't need the command args anymore
	allOtherArgs := make([]string, 0)
	commandNameArgs := make([]string, 0)
	for i := 0; i < len(args); i++ {
		if exists(commandNameIndexes, i) {
			commandNameArgs = append(commandNameArgs, args[i])
			continue
		}
		allOtherArgs = append(allOtherArgs, args[i])
	}

	return command, commandNameArgs, allOtherArgs, nil
}

func exists(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}

	return false
}

func (c *app) matchCommandByName(arg string, commands []Command[any]) Command[any] {
	var command Command[any]
	for i := 0; i < len(commands); i++ {
		cmd := commands[i]
		if cmd.Name("") == arg {
			command = cmd
			break
		}
	}

	return command
}
