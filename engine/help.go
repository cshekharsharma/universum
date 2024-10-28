package engine

import (
	"fmt"
	"universum/config"
)

func getGenericHelpContent() string {
	return fmt.Sprintf("%s is a fast key-value pair in-memory database, "+
		"that supports periodic implicite backups and data reconstruction.\n\n"+
		"It uses RESP3 format for client-server communication.\n\n"+
		"For help related to a specific command, please use: `HELP <command>\n\n`", config.AppNameLabel)
}

func getCommandHelpContent(command string) string {
	switch command {

	case CommandPing:
		return "USAGE:\n\n\tPING\n"

	case CommandExists:
		return "USAGE:\n\n\tEXISTS <key:string>\n"

	case CommandGet:
		return "USAGE:\n\n\tGET <key:string>\n"

	case CommandSet:
		return "USAGE:\n\n\tEXISTS <key:string> <value:any> <ttl:int>\n"

	case CommandDelete:
		return "USAGE:\n\n\tDELETE <key:string>\n"

	case CommandIncr:
		return "USAGE:\n\n\tINCR <key:string> <value:int>\n"

	case CommandDecr:
		return "USAGE:\n\n\tDECR <key:string> <value:int>\n"

	case CommandAppend:
		return "USAGE:\n\n\tAPPEND <key:string> <value:string>\n"

	case CommandMGet:
		return "USAGE:\n\n\tMGET <keys:[]string>\n"

	case CommandMSet:
		return "USAGE:\n\n\tMSET <KvMap:map[string][any]>\n"

	case CommandMDelete:
		return "USAGE:\n\n\tMDELETE <keys:[]string>\n"

	case CommandTTL:
		return "USAGE:\n\n\tTTL <key:string>\n"

	case CommandExpire:
		return "USAGE:\n\n\tEXPIRE <key:string> <ttl:int>\n"

	case CommandSnapshot:
		return "USAGE:\n\n\tSNAPSHOT\n"

	case CommandInfo:
		return "USAGE:\n\n\tINFO\n"

	default:
		return fmt.Sprintf("\nInvalid subcommand `%s`. Retry with correct subcommand\n", command)
	}
}
