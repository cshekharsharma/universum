package engine

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"universum/engine/entity"
	"universum/resp3"
)

const (
	COMMAND_GET    string = "GET"
	COMMAND_SET    string = "SET"
	COMMAND_DELETE string = "DELETE"
	COMMAND_EXISTS string = "EXISTS"
	COMMAND_INCR   string = "INCR"
	COMMAND_DECR   string = "DECR"
)

func ExecuteCommand(buffer *bufio.Reader) (string, error) {
	command, err := parseCommand(buffer)

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

func parseCommand(buffer *bufio.Reader) (*entity.Command, error) {
	decodedResp, err := resp3.Decode(buffer)

	if err != nil {
		return nil, err
	}

	return getCommandFromRESP(decodedResp)
}

func getCommandFromRESP(decodedResp interface{}) (*entity.Command, error) {
	decodedList, ok := decodedResp.([]interface{})

	if !ok {
		return nil, errors.New("incompatible RESP3 input, expected a list")
	}

	rawcommand, readerr := resp3.Encode(decodedResp)
	if readerr != nil {
		return nil, errors.New("failed to read the decoded RESP command")
	}

	command := &entity.Command{
		Name: strings.ToUpper(fmt.Sprintf("%v", decodedList[0])),
		Args: decodedList[1:],
		Raw:  []byte(rawcommand),
	}

	return command, nil
}

func executeCommand(command *entity.Command) (string, error) {
	switch command.Name {
	case COMMAND_EXISTS:
		return executeEXISTS(command), nil

	case COMMAND_SET:
		return executeSET(command), nil

	case COMMAND_GET:
		return executeGET(command), nil

	case COMMAND_DELETE:
		return executeDELETE(command), nil

	case COMMAND_INCR:
		return executeINCRDECR(command, true), nil

	case COMMAND_DECR:
		return executeINCRDECR(command, false), nil

	default:
		return "", fmt.Errorf("invalid command `%s` provided", command)
	}

}
