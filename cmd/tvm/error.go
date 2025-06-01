package tvm

import (
	"fmt"
)

// InvalidOutputParamErr indicates that the output parameter for the provided operation is
// invalid---that is, the parameter is not in address format
type InvalidOutputParamErr struct {
	Operation string
}

func (i InvalidOutputParamErr) Error() string {
	return fmt.Sprintf("invalid output parameter for %v operation", i.Operation)
}
