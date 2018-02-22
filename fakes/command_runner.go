package fakes

type CommandRunner struct {
	RunCall struct {
		CallCount int
		Receives  struct {
			Command string
			Args    []string
		}
		Returns struct {
			Error error
		}
	}
}

func (c *CommandRunner) Run(command string, args []string) error {
	c.RunCall.CallCount++
	c.RunCall.Receives.Command = command
	c.RunCall.Receives.Args = args
	return c.RunCall.Returns.Error
}
