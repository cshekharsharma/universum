package engine

import (
	"bufio"
	"errors"
	"strings"
)

const (
	COMMAND_GET    string = "GET"
	COMMAND_SET    string = "SET"
	COMMAND_DELETE string = "DELETE"
	COMMAND_EXISTS string = "EXISTS"
)

type Command struct {
	Name string
	Args []interface{}
}

func ExecuteCommand(buff *bufio.Reader) (any, error) {
	command, err := parseCommand(buff)

	if err != nil {
		return nil, err
	}

	output, err := execute(command)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func parseCommand(buff *bufio.Reader) (*Command, error) {
	raw, err := decodeRESP3(buff)

	if err != nil {
		return nil, err
	}

	parsedCommand := raw.([]interface{})

	command := &Command{
		Name: strings.ToUpper(parsedCommand[0].(string)),
		Args: parsedCommand[1:],
	}

	return command, nil
}

func execute(command *Command) (any, error) {
	switch command.Name {
	case "PING":
		return Op_Get(command.Args)
	case "SET":
		return Op_Set(command.Args)
	default:
		return nil, errors.New("Invalid command provided")
	}

	return nil, nil
}
