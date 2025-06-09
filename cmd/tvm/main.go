package main

import (
	tvm "tvm/internal/virtual_machine"
)

func main() {
	machine := tvm.NewTsvetokVirtualMachine([]int{9})
	machine.Execute()
}
