package core

type Command struct {
	Name      string	`json:"name"`
	Arguments []string	`json:"args"`
}

type State struct {
	Properties map[string]string
}

type SimpleDevice interface {
	Id() string
	ListCommands() ([]Command, error)
	SendCommand(command *Command) error
	GetState() *State
	SubcribeToStateChanges() (chan *State, error)
	SubcribeToErrorChanges() (chan error, error)
}
