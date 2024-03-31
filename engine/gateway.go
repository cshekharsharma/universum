package engine

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"universum/engine/entity"
	"universum/utils"
)

const (
	COMMAND_GET    string = "GET"
	COMMAND_SET    string = "SET"
	COMMAND_DELETE string = "DELETE"
	COMMAND_EXISTS string = "EXISTS"
)

func ExecuteCommand(buff *bufio.Reader) (string, error) {
	command, err := parseCommand(buff)

	if err != nil {
		return "", err
	}

	fmt.Printf("REQUEST: %#v\n", command)
	output, err := executeCommand(command)

	if err != nil {
		return "", err
	}

	return output, nil
}

func parseCommand(buff *bufio.Reader) (*entity.Command, error) {
	raw, err := utils.DecodeRESP3(buff)

	if err != nil {
		return nil, err
	}

	parsedCommand, ok := raw.([]interface{})

	if !ok {
		return nil, errors.New("incompatible RESP3 input, expected a list")
	}

	command := &entity.Command{
		Name: strings.ToUpper(fmt.Sprintf("%v", parsedCommand[0])),
		Args: parsedCommand[1:],
	}

	return command, nil
}

func executeCommand(command *entity.Command) (string, error) {
	switch command.Name {
	case COMMAND_GET:
		return executeGET(command), nil

	case COMMAND_SET:
		return executeSET(command), nil

	default:
		return "", fmt.Errorf("invalid command `%s` provided", command)
	}

}
