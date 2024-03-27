// lists and implements all operations supported by the database
package engine

func Shutdown() {
	// do all the shut down operations, such as fsyncing AOF
	// and freeing up occupied resources and memory.
}

func Op_Get(args []interface{}) (any, error)    { return nil, nil }
func Op_Set(args []interface{}) (any, error)    { return nil, nil }
func Op_Delete(args []interface{}) (any, error) { return nil, nil }
func Op_Exists(args []interface{}) (any, error) { return nil, nil }
