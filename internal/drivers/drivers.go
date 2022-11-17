package drivers

import (
	"cuelang.org/go/cue"
	"devopzilla.com/guku/internal/stack"
)

type Driver interface {
	match(resource cue.Value) bool
	ApplyAll(stack *stack.Stack) error
}
