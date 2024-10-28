package engine

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"universum/entity"
	"universum/internal/logger"
	"universum/resp3"
)

const (
	CommandPing     string = "PING"
	CommandExists   string = "EXISTS"
	CommandGet      string = "GET"
	CommandSet      string = "SET"
	CommandDelete   string = "DELETE"
	CommandIncr     string = "INCR"
	CommandDecr     string = "DECR"
	CommandAppend   string = "APPEND"
	CommandMGet     string = "MGET"
	CommandMSet     string = "MSET"
	CommandMDelete  string = "MDELETE"
	CommandTTL      string = "TTL"
	CommandExpire   string = "EXPIRE"
	CommandSnapshot string = "SNAPSHOT"
	CommandInfo     string = "INFO"
	CommandHelp     string = "HELP"
)

func ExecuteCommand(buffer *bufio.Reader, timeout time.Duration) (string, error) {
	command, err := parseCommand(buffer)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	defer cancel()

	logger.Get().Debug("REQUEST: %#v", command)
	output, err := executeCommand(ctx, command)
	logger.Get().Debug("RESPONSE: %#v", output)

	if err != nil {
		return "", err
	}

	AddCommandsProcessed(1)
	return output, nil
}

func parseCommand(buffer *bufio.Reader) (*entity.Command, error) {
	buffer.Peek(1)
	AddNetworkBytesReceived(int64(buffer.Buffered()))
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

func executeCommand(ctx context.Context, command *entity.Command) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		// Continue processing the command
	}

	switch command.Name {
	case CommandPing:
		return executePING(command), nil

	case CommandExists:
		return executeEXISTS(command), nil

	case CommandGet:
		return executeGET(command), nil

	case CommandSet:
		return executeSET(command), nil

	case CommandDelete:
		return executeDELETE(command), nil

	case CommandIncr:
		return executeINCRDECR(command, true), nil

	case CommandDecr:
		return executeINCRDECR(command, false), nil

	case CommandAppend:
		return executeAPPEND(command), nil

	case CommandMGet:
		return executeMGET(command), nil

	case CommandMSet:
		return executeMSET(command), nil

	case CommandMDelete:
		return executeMDELETE(command), nil

	case CommandTTL:
		return executeTTL(command), nil

	case CommandExpire:
		return executeEXPIRE(command), nil

	case CommandSnapshot:
		return executeSNAPSHOT(command), nil

	case CommandInfo:
		return executeINFO(command), nil

	case CommandHelp:
		return executeHELP(command), nil

	default:
		return "", fmt.Errorf("invalid command `%s` provided", command.Name)
	}

}
