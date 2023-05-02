package types

type Action string

const (
	AddItem    Action = "add"
	RemoveItem Action = "remove"
	GetItem    Action = "get"
	GetAll     Action = "getall"
)

func (a Action) String() string {
	return string(a)
}

type Message struct {
	Action Action `json:"action"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}
