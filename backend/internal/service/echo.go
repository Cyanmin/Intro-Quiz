package service

// EchoService simply echoes back received messages.
type EchoService struct{}

// NewEchoService creates a new EchoService.
func NewEchoService() *EchoService {
	return &EchoService{}
}

// ProcessMessage returns the message as-is.
func (e *EchoService) ProcessMessage(mt int, msg []byte) (int, []byte) {
	return mt, msg
}
