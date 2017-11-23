package listener

import (
	"github.com/davidwalter0/forwarder/tracer"
)

// Equal compares two ManagedListener objects
func (lhs *ManagedListener) Equal(rhs *ManagedListener) bool {
	defer trace.Tracer.Detailed(trace.Detail).Enable(trace.Enabled).ScopedTrace()()
	return lhs.Source == rhs.Source &&
		lhs.Sink == rhs.Sink &&
		lhs.EnableEp == rhs.EnableEp &&
		lhs.Service == rhs.Service &&
		lhs.Namespace == rhs.Namespace
}

// Copy points w/o erasing EndPoints
func (lhs *ManagedListener) Copy(rhs *ManagedListener) *ManagedListener {
	lhs.Source = rhs.Source
	lhs.Sink = rhs.Sink
	lhs.EnableEp = rhs.EnableEp
	lhs.Service = rhs.Service
	lhs.Namespace = rhs.Namespace
	return lhs
}

// Equal compares two PipeDefinition objects
func (lhs *PipeDefinition) Equal(rhs *PipeDefinition) bool {
	defer trace.Tracer.Detailed(trace.Detail).Enable(trace.Enabled).ScopedTrace()()
	// if lhs == nil {
	// 	lhs = new(PipeDefinition)
	// }
	return lhs.Source == rhs.Source &&
		lhs.Sink == rhs.Sink &&
		lhs.EnableEp == rhs.EnableEp &&
		lhs.Service == rhs.Service &&
		lhs.Namespace == rhs.Namespace
}

// Copy points w/o erasing EndPoints
func (lhs *PipeDefinition) Copy(rhs *PipeDefinition) *PipeDefinition {
	// if lhs == nil {
	// 	lhs = new(PipeDefinition)
	// }
	lhs.Source = rhs.Source
	lhs.Sink = rhs.Sink
	lhs.EnableEp = rhs.EnableEp
	lhs.Service = rhs.Service
	lhs.Namespace = rhs.Namespace
	return lhs
}
