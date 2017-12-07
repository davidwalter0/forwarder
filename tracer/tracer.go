package trace

import (
	"github.com/davidwalter0/go-tracer"
)

var Tracer = tracer.New()
var Detail = false
var Enabled = false

func init() {
	Tracer.Detailed(Detail).Enable(Enabled)
}
