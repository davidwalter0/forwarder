package trace

import (
	"github.com/davidwalter0/go-tracer"
)

var Tracer = tracer.New()
var Detail = true
var Enabled = true

func init() {
	Tracer.Detailed(Detail).Enable(Enabled)
}
