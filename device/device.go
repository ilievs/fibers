package device

type Command struct {
	Name      string
	Arguments []string
}

type State struct {
	Properties map[string]string
}

type SimpleDevice interface {
	Id() string
	ListCommands() ([]Command, error)
	SendCommand(command Command) error
	GetState() State
	SubcribeToStateChanges() (chan State, error)
	SubcribeToErrorChanges() (chan error, error)
}
