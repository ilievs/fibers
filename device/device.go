package device

type Command struct {
	Name      string
	Arguments []string
}

type State struct {
	Properties map[string]string
}

type SimpleDevice interface {
	ListCommands() ([]Command, error)
	SendCommand(command Command) error
	SubcribeToStateChanges() (chan State, error)
	SubcribeToErrorChanges() (chan error, error)
}
