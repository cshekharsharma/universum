package resp3

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

type testCase struct {
	Description       string
	Input             interface{}
	ExpectedEncOutput string
	ExpectedDecOutput interface{}
}

var testCases = []testCase{
	{
		Description:       "Simple String",
		Input:             "hello",
		ExpectedEncOutput: "$5\r\nhello\r\n",
		ExpectedDecOutput: "hello",
	},
	{
		Description:       "Bulk String",
		Input:             "bulk string example",
		ExpectedEncOutput: "$19\r\nbulk string example\r\n",
		ExpectedDecOutput: "bulk string example",
	},
	{
		Description:       "Integer",
		Input:             42,
		ExpectedEncOutput: ":42\r\n",
		ExpectedDecOutput: 42,
	},
	{
		Description:       "Boolean (true)",
		Input:             true,
		ExpectedEncOutput: "#t\r\n",
		ExpectedDecOutput: true,
	},
	{
		Description:       "Boolean (false)",
		Input:             false,
		ExpectedEncOutput: "#f\r\n",
		ExpectedDecOutput: false,
	},
	{
		Description:       "Nil",
		Input:             nil,
		ExpectedEncOutput: "_\r\n",
		ExpectedDecOutput: nil,
	},
	{
		Description:       "Array",
		Input:             []interface{}{"msg", 123},
		ExpectedEncOutput: "*2\r\n$3\r\nmsg\r\n:123\r\n",
		ExpectedDecOutput: []interface{}{"msg", 123},
	},
	{
		Description:       "Array of Interfaces",
		Input:             []interface{}{true, nil, "test"},
		ExpectedEncOutput: "*3\r\n#t\r\n_\r\n$4\r\ntest\r\n",
		ExpectedDecOutput: []interface{}{true, nil, "test"},
	},
	{
		Description: "Map map[string]interface{}",
		Input: map[string]interface{}{
			"age":       30,
			"isStudent": false,
			"grades": map[string]interface{}{
				"math":    95,
				"science": 90,
			},
		},
		ExpectedEncOutput: "%6\r\n+age\r\n:30\r\n+isStudent\r\n#f\r\n+grades\r\n%4\r\n+math\r\n:95\r\n+science\r\n:90\r\n",
		ExpectedDecOutput: map[string]interface{}{
			"age":       30,
			"isStudent": false,
			"grades": map[string]interface{}{
				"math":    95,
				"science": 90,
			},
		},
	},
}

func TestEncode(t *testing.T) {

	for _, tc := range testCases {
		actualEncOutput, err := Encode(tc.Input)

		if err != nil {
			t.Errorf("%s: error while encoding %v", tc.Description, tc.Input)
		}

		if actualEncOutput != tc.ExpectedEncOutput {
			t.Errorf("%s: Encoding assertion failed [%v != %v", tc.Description, actualEncOutput, tc.ExpectedEncOutput)
		}

	}
}

func TestDecode(t *testing.T) {

	for _, tc := range testCases {
		actualDecOutput, err := Decode(bufio.NewReader(strings.NewReader(tc.ExpectedEncOutput)))

		if err != nil {
			t.Errorf("%s: error while decoding %v", tc.Description, tc.ExpectedEncOutput)
		}

		fActual := fmt.Sprintf("%#v", actualDecOutput)
		fExpected := fmt.Sprintf("%#v", tc.ExpectedDecOutput)

		if fActual != fExpected {
			t.Errorf("%s: Decoding assertion failed for [ %v != %v ]", tc.Description, fActual, fExpected)
		}
	}
}
