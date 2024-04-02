package entity

type Command struct {
	Name string
	Args []interface{}
	Raw  []byte
}
