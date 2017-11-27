package listener

import (
	"github.com/davidwalter0/forwarder/tracer"
)

// Equal compares two ManagedListener objects
func (lhs *ManagedListener) Equal(rhs *ManagedListener) bool {
	defer trace.Tracer.ScopedTrace()()
	return lhs.Name == rhs.Name &&
		lhs.Source == rhs.Source &&
		lhs.Sink == rhs.Sink &&
		lhs.EnableEp == rhs.EnableEp &&
		lhs.Service == rhs.Service &&
		lhs.Namespace == rhs.Namespace &&
		lhs.Debug == rhs.Debug
}

// Copy points w/o erasing EndPoints
func (lhs *ManagedListener) Copy(rhs *ManagedListener) *ManagedListener {
	lhs.Name = rhs.Name
	lhs.Source = rhs.Source
	lhs.Sink = rhs.Sink
	lhs.EnableEp = rhs.EnableEp
	lhs.Service = rhs.Service
	lhs.Namespace = rhs.Namespace
	lhs.Debug = rhs.Debug
	lhs.Active = rhs.Active
	return lhs
}

// Equal compares two PipeDefinition objects
func (lhs *PipeDefinition) Equal(rhs *PipeDefinition) bool {
	defer trace.Tracer.ScopedTrace()()
	return lhs.Name == rhs.Name &&
		lhs.Source == rhs.Source &&
		lhs.Sink == rhs.Sink &&
		lhs.EnableEp == rhs.EnableEp &&
		lhs.Service == rhs.Service &&
		lhs.Namespace == rhs.Namespace &&
		lhs.Debug == rhs.Debug
}

// Copy points w/o erasing EndPoints
func (lhs *PipeDefinition) Copy(rhs *PipeDefinition) *PipeDefinition {
	lhs.Name = rhs.Name
	lhs.Source = rhs.Source
	lhs.Sink = rhs.Sink
	lhs.EnableEp = rhs.EnableEp
	lhs.Service = rhs.Service
	lhs.Namespace = rhs.Namespace
	lhs.Debug = rhs.Debug
	return lhs
}
