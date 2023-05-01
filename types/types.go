package types

type Action string

const (
	AddItem    Action = "AddItem"
	RemoveItem Action = "RemoveItem"
	GetItem    Action = "GetItem"
	GetAll     Action = "GetAll"
)

func (a Action) String() string {
	return string(a)
}

type Message struct {
	Action Action
	Key    string
	Value  string
}
