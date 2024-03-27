package engine

import "bufio"

type Command struct {
	Name     string
	Argument string
}

func ExecuteCommand(buff *bufio.Reader) (any, error) {
	return parseCommand(buff)
}

func parseCommand(buff *bufio.Reader) (interface{}, error) {
	parsed, err := decodeRESP3(buff)
	return parsed, err
}
