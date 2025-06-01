package tvm

import (
	"fmt"
)

type InvalidOutputParamErr struct {
	Operation string
}

func (i InvalidOutputParamErr) Error() string {
	return fmt.Sprintf("invalid output parameter for %v operation", i.Operation)
}
