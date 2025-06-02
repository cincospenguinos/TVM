package virtual_machine

import (
	"fmt"
)

// AttemptedLastAddressWriteErr indicates that an operation attempted to write to the last address
// register, which is write protected. Note that the only operation that may write to that register
// is the jit instruction, which ignores this error
type AttemptedLastAddressWriteErr struct{}

func (_ AttemptedLastAddressWriteErr) Error() string {
	return fmt.Sprintf("attempted to write to last address register '%v'", RegisterLastAddress)
}

// InvalidOutputParamErr indicates that the output parameter for the provided operation is
// invalid---that is, the parameter is not in address format
type InvalidOutputParamErr struct {
	Operation string
}

func (i InvalidOutputParamErr) Error() string {
	return fmt.Sprintf("invalid output parameter for %v operation", i.Operation)
}
