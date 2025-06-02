package virtual_machine

import ()

type MockInputInterface struct {
	NumberToReturn int
}

func (m MockInputInterface) ReceiveInput() int {
	return m.NumberToReturn
}

type MockOutputInterface struct {
	LastNumberReceived *int
}

func (m *MockOutputInterface) EmitOutput(number int) {
	m.LastNumberReceived = &number
}
