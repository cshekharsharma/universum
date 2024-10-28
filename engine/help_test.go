package engine

import (
	"fmt"
	"testing"
)

func TestGetGenericHelpContent(t *testing.T) {
	result := getGenericHelpContent()

	if len(result) == 0 {
		t.Errorf("Expected non-empty help content, got empty")
	}
}

func TestGetCommandHelpContent(t *testing.T) {
	testCases := []struct {
		command  string
		expected string
	}{
		{CommandPing, "USAGE:\n\n\tPING\n"},
		{CommandExists, "USAGE:\n\n\tEXISTS <key:string>\n"},
		{CommandGet, "USAGE:\n\n\tGET <key:string>\n"},
		{CommandSet, "USAGE:\n\n\tEXISTS <key:string> <value:any> <ttl:int>\n"},
		{CommandDelete, "USAGE:\n\n\tDELETE <key:string>\n"},
		{CommandIncr, "USAGE:\n\n\tINCR <key:string> <value:int>\n"},
		{CommandDecr, "USAGE:\n\n\tDECR <key:string> <value:int>\n"},
		{CommandAppend, "USAGE:\n\n\tAPPEND <key:string> <value:string>\n"},
		{CommandMGet, "USAGE:\n\n\tMGET <keys:[]string>\n"},
		{CommandMSet, "USAGE:\n\n\tMSET <KvMap:map[string][any]>\n"},
		{CommandMDelete, "USAGE:\n\n\tMDELETE <keys:[]string>\n"},
		{CommandTTL, "USAGE:\n\n\tTTL <key:string>\n"},
		{CommandExpire, "USAGE:\n\n\tEXPIRE <key:string> <ttl:int>\n"},
		{CommandSnapshot, "USAGE:\n\n\tSNAPSHOT\n"},
		{CommandInfo, "USAGE:\n\n\tINFO\n"},
		{"InvalidCommand", "\nInvalid subcommand `InvalidCommand`. Retry with correct subcommand\n"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("HelpContent_%s", tc.command), func(t *testing.T) {
			result := getCommandHelpContent(tc.command)
			if result != tc.expected {
				t.Errorf("For command '%s', expected:\n%s\nGot:\n%s", tc.command, tc.expected, result)
			}
		})
	}
}
