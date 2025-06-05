package virtual_machine

import ()

// InputInterface represents any external service or program that can be called upon for
// an integer
type InputInterface interface {
	// ReceiveInput acquires and returns an integer from an external source
	ReceiveInput() int
}

// OutputInterface represents any external service or program that an integer can be
// emitted
type OutputInterface interface {
	// EmitOutput emits the integer provided to a given target
	EmitOutput(int)
}
