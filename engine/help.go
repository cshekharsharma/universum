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

	case COMMAND_PING:
		return "USAGE:\n\n\tPING\n"

	case COMMAND_EXISTS:
		return "USAGE:\n\n\tEXISTS <key:string>\n"

	case COMMAND_GET:
		return "USAGE:\n\n\tGET <key:string>\n"

	case COMMAND_SET:
		return "USAGE:\n\n\tEXISTS <key:string> <value:any> <ttl:int>\n"

	case COMMAND_DELETE:
		return "USAGE:\n\n\tDELETE <key:string>\n"

	case COMMAND_INCR:
		return "USAGE:\n\n\tINCR <key:string> <value:int>\n"

	case COMMAND_DECR:
		return "USAGE:\n\n\tDECR <key:string> <value:int>\n"

	case COMMAND_APPEND:
		return "USAGE:\n\n\tAPPEND <key:string> <value:string>\n"

	case COMMAND_MGET:
		return "USAGE:\n\n\tMGET <keys:[]string>\n"

	case COMMAND_MSET:
		return "USAGE:\n\n\tMSET <KvMap:map[string][any]>\n"

	case COMMAND_MDELETE:
		return "USAGE:\n\n\tMDELETE <keys:[]string>\n"

	case COMMAND_TTL:
		return "USAGE:\n\n\tTTL <key:string>\n"

	case COMMAND_EXPIRE:
		return "USAGE:\n\n\tEXPIRE <key:string> <ttl:int>\n"

	case COMMAND_SNAPSHOT:
		return "USAGE:\n\n\tSNAPSHOT\n"

	case COMMAND_INFO:
		return "USAGE:\n\n\tINFO\n"

	default:
		return fmt.Sprintf("\nInvalid subcommand `%s`. Retry with correct subcommand\n", command)
	}
}
