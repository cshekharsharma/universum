package utils

import (
	"reflect"
	"testing"
	"universum/entity"
)

func TestValidateArguments(t *testing.T) {
	tests := []struct {
		name         string
		cmd          *entity.Command
		validations  []ValidationRule
		wantValid    bool
		wantResponse []interface{}
	}{
		{
			name: "CorrectArgs",
			cmd: &entity.Command{
				Args: []interface{}{"hello", 123},
			},
			validations: []ValidationRule{
				{Name: "greeting", Datatype: reflect.String},
				{Name: "number", Datatype: reflect.Int},
			},
			wantValid:    true,
			wantResponse: []interface{}{},
		},
		{
			name: "IncorrectArgs",
			cmd: &entity.Command{
				Args: []interface{}{"hello", "world"},
			},
			validations: []ValidationRule{
				{Name: "greeting", Datatype: reflect.String},
				{Name: "number", Datatype: reflect.Int},
			},
			wantValid: false,
			wantResponse: []interface{}{
				nil,
				entity.CRC_INVALID_CMD_INPUT,
				"ERR: number has invalid type. int expected",
			},
		},
		{
			name: "IncorrectArgCount",
			cmd: &entity.Command{
				Args: []interface{}{"hello"},
			},
			validations: []ValidationRule{
				{Name: "greeting", Datatype: reflect.String},
				{Name: "number", Datatype: reflect.Int},
			},
			wantValid: false,
			wantResponse: []interface{}{
				nil,
				entity.CRC_INVALID_CMD_INPUT,
				"ERR: Incorrect number of arguments provided. Want=2, Have=1",
			},
		},
		{
			name: "WildcardDatatype",
			cmd: &entity.Command{
				Args: []interface{}{"hello", 123},
			},
			validations: []ValidationRule{
				{Name: "greeting", Datatype: reflect.Interface},
				{Name: "number", Datatype: reflect.Interface},
			},
			wantValid:    true,
			wantResponse: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, response := ValidateArguments(tt.cmd, tt.validations)
			if valid != tt.wantValid {
				t.Errorf("ValidateArguments() valid = %v, want %v", valid, tt.wantValid)
			}
			if !reflect.DeepEqual(response, tt.wantResponse) {
				t.Errorf("ValidateArguments() response = %v, want %v", response, tt.wantResponse)
			}
		})
	}
}
