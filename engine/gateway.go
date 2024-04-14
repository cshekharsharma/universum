package engine

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"universum/engine/entity"
	"universum/internal/logger"
	"universum/resp3"
)

const (
	COMMAND_PING     string = "PING"
	COMMAND_EXISTS   string = "EXISTS"
	COMMAND_GET      string = "GET"
	COMMAND_SET      string = "SET"
	COMMAND_DELETE   string = "DELETE"
	COMMAND_INCR     string = "INCR"
	COMMAND_DECR     string = "DECR"
	COMMAND_APPEND   string = "APPEND"
	COMMAND_MGET     string = "MGET"
	COMMAND_MSET     string = "MSET"
	COMMAND_MDELETE  string = "MDELETE"
	COMMAND_TTL      string = "TTL"
	COMMAND_EXPIRE   string = "EXPIRE"
	COMMAND_SNAPSHOT string = "SNAPSHOT"
	COMMAND_INFO     string = "INFO"
	COMMAND_HELP     string = "HELP"
)

func ExecuteCommand(buffer *bufio.Reader) (string, error) {
	command, err := parseCommand(buffer)

	if err != nil {
		return "", err
	}

	fmt.Printf("REQUEST: %#v\n", command)
	output, err := executeCommand(command)
	fmt.Printf("RESPONSE: %#v\n", output)

	if err != nil {
		return "", err
	}

	AddCommandsProcessed(1)
	return output, nil
}

func parseCommand(buffer *bufio.Reader) (*entity.Command, error) {
	buffer.Peek(1)
	AddNetworkBytesReceived(int64(buffer.Buffered()))
	logger.Get().Info("Buffered: %d", buffer.Buffered())
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

	command := &entity.Command{
		Name: strings.ToUpper(fmt.Sprintf("%v", decodedList[0])),
		Args: decodedList[1:],
	}

	return command, nil
}

func executeCommand(command *entity.Command) (string, error) {
	switch command.Name {

	case COMMAND_PING:
		return executePING(command), nil

	case COMMAND_EXISTS:
		return executeEXISTS(command), nil

	case COMMAND_GET:
		return executeGET(command), nil

	case COMMAND_SET:
		return executeSET(command), nil

	case COMMAND_DELETE:
		return executeDELETE(command), nil

	case COMMAND_INCR:
		return executeINCRDECR(command, true), nil

	case COMMAND_DECR:
		return executeINCRDECR(command, false), nil

	case COMMAND_APPEND:
		return executeAPPEND(command), nil

	case COMMAND_MGET:
		return executeMGET(command), nil

	case COMMAND_MSET:
		return executeMSET(command), nil

	case COMMAND_MDELETE:
		return executeMDELETE(command), nil

	case COMMAND_TTL:
		return executeTTL(command), nil

	case COMMAND_EXPIRE:
		return executeEXPIRE(command), nil

	case COMMAND_SNAPSHOT:
		return executeSNAPSHOT(command), nil

	case COMMAND_INFO:
		return executeINFO(command), nil

	case COMMAND_HELP:
		return executeHELP(command), nil

	default:
		return "", fmt.Errorf("invalid command `%s` provided", command.Name)
	}

}
