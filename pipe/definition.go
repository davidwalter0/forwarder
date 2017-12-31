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
	Key       string `json:"key"       help:"map key (ns/name k8s services|yaml name for definition)"`
	Source    string `json:"source"    help:"source ingress point host:port"`
	Sink      string `json:"sink"      help:"sink service point   host:port"`
	Endpoints EP     `json:"endpoints" help:"endpoints (sinks) k8s api / config"`
	EnableEp  bool   `json:"enable-ep" help:"enable endpoints from service"`
	Name      string `json:"service"   help:"service name"`
	Namespace string `json:"namespace" help:"service namespace"`
	Mode      string `json:"mode"      help:"mode of use for this service"`
	Debug     bool   `json:"debug"     help:"enable debug for this pipe"`
	InCluster bool   `json:"incluster" help:"incluster forwarding vs external forward to NodePort(s)"`
}

// NewFromDefinition create and initialize a Definition
func NewFromDefinition(pipe *Definition) (p *Definition) {
	if pipe != nil {
		defer trace.Tracer.ScopedTrace()()
		p = &Definition{
			// Name is the key of yaml map
			Key:       pipe.Key,
			Source:    pipe.Source,
			Sink:      pipe.Sink,
			EnableEp:  pipe.EnableEp,
			Name:      pipe.Name,
			Namespace: pipe.Namespace,
			Debug:     pipe.Debug,
			Mode:      pipe.Mode,
			InCluster: true,
		}
	}
	return
}

// Definitions from text description in yaml
type Definitions map[string]*Definition

// Equal compares two pipe.Definition objects
func (lhs *Definition) Equal(rhs *Definition) bool {
	defer trace.Tracer.ScopedTrace()()
	return lhs.Key == rhs.Key &&
		lhs.Source == rhs.Source &&
		lhs.Sink == rhs.Sink &&
		lhs.EnableEp == rhs.EnableEp &&
		lhs.Name == rhs.Name &&
		lhs.Namespace == rhs.Namespace &&
		lhs.Mode == rhs.Mode &&
		lhs.Debug == rhs.Debug &&
		lhs.InCluster == rhs.InCluster
}

// Copy points w/o erasing EndPoints
func (lhs *Definition) Copy(rhs *Definition) *Definition {
	lhs.Key = rhs.Key
	lhs.Source = rhs.Source
	lhs.Sink = rhs.Sink
	lhs.EnableEp = rhs.EnableEp
	lhs.Name = rhs.Name
	lhs.Namespace = rhs.Namespace
	lhs.Mode = rhs.Mode
	lhs.Debug = rhs.Debug
	lhs.InCluster = rhs.InCluster
	return lhs
}
