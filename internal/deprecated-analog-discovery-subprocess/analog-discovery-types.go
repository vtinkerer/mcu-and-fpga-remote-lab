package analogdiscoverysubprocess

type Command struct {
	Content string `json:"content"`
}

type ProcessRunner interface {
	Start() error
	ReadOutput(chan<- string)
	WriteInput(string) error
	ReadError(chan<- string)
	Stop() error
}