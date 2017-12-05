package pipe

import (
	"sort"

	"github.com/davidwalter0/forwarder/tracer"
)

// EP slice of endpoints
type EP []string

// Equal compare two endpoint arrays for equality
func (ep *EP) Equal(rhs *EP) (rc bool) {
	if ep != nil && rhs != nil && len(*ep) == len(*rhs) {
		sort.Strings(*ep)
		sort.Strings(*rhs)
		for i, v := range *ep {
			if v != (*rhs)[i] {
				return
			}
		}
	} else {
		return
	}
	return true
}

// Definition maps source to sink
type Definition struct {
	Name      string `json:"name"      help:"map key"`
	Source    string `json:"source"    help:"source ingress point host:port"`
	Sink      string `json:"sink"      help:"sink service point   host:port"`
	Endpoints *EP    `json:"endpoints" help:"endpoints (sinks) k8s api / config"`
	EnableEp  bool   `json:"enable-ep" help:"enable endpoints from service"`
	Service   string `json:"service"   help:"service name"`
	Namespace string `json:"namespace" help:"service namespace"`
	Mode      string `json:"mode"      help:"mode of use for this service"`
	Debug     bool   `json:"debug"     help:"enable debug for this pipe"`
}

// NewFromDefinition create and initialize a Definition
func NewFromDefinition(pipe *Definition) (p *Definition) {
	if pipe != nil {
		defer trace.Tracer.ScopedTrace()()
		p = &Definition{
			// Name is the key of yaml map
			// Name:      name,
			Source:    pipe.Source,
			Sink:      pipe.Sink,
			EnableEp:  pipe.EnableEp,
			Service:   pipe.Service,
			Namespace: pipe.Namespace,
			Debug:     pipe.Debug,
			Mode:      pipe.Mode,
		}
	}
	return
}

// Definitions from text description in yaml
type Definitions map[string]*Definition

// Equal compares two pipe.Definition objects
func (lhs *Definition) Equal(rhs *Definition) bool {
	defer trace.Tracer.ScopedTrace()()
	return lhs.Name == rhs.Name &&
		lhs.Source == rhs.Source &&
		lhs.Sink == rhs.Sink &&
		lhs.EnableEp == rhs.EnableEp &&
		lhs.Service == rhs.Service &&
		lhs.Namespace == rhs.Namespace &&
		lhs.Mode == rhs.Mode &&
		lhs.Debug == rhs.Debug
}

// Copy points w/o erasing EndPoints
func (lhs *Definition) Copy(rhs *Definition) *Definition {
	lhs.Name = rhs.Name
	lhs.Source = rhs.Source
	lhs.Sink = rhs.Sink
	lhs.EnableEp = rhs.EnableEp
	lhs.Service = rhs.Service
	lhs.Namespace = rhs.Namespace
	lhs.Mode = rhs.Mode
	lhs.Debug = rhs.Debug
	return lhs
}
