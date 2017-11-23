package listener

import (
	"fmt"
	"testing"
)

// func (lhs *ManagedListener) Equal(rhs *ManagedListener) bool {

// // PipeDefinition maps source to sink
// type PipeDefinition struct {
// 	Source    string   `json:"source"    help:"source ingress point host:port"`
// 	Sink      string   `json:"sink"      help:"sink service point   host:port"`
// 	EndPoints []string `json:"endpoints" help:"endpoints (sinks) k8s api / config"`
// 	EnableEp  bool     `json:"enable-ep" help:"enable endpoints from service"`
// 	Service   string   `json:"service"   help:"service name"`
// 	Namespace string   `json:"namespace" help:"service namespace"`
// }
type TestPipe map[string][]PipeDefinition

var _testPipe1 = TestPipe{
	"Equal": []PipeDefinition{
		PipeDefinition{
			Source:    "0.0.0.0:8001",
			Sink:      "0.0.0.0:8002",
			EndPoints: []string{},
			EnableEp:  false,
			Service:   "echo",
			Namespace: "test",
		},
		PipeDefinition{
			Source:    "0.0.0.0:8001",
			Sink:      "0.0.0.0:8002",
			EndPoints: []string{},
			EnableEp:  false,
			Service:   "echo",
			Namespace: "test",
		},
	},
	"!Equal": []PipeDefinition{
		PipeDefinition{
			Source:    "0.0.0.0:8002",
			Sink:      "0.0.0.0:8003",
			EndPoints: []string{},
			EnableEp:  false,
			Service:   "echo",
			Namespace: "test",
		},
		PipeDefinition{
			Source:    "0.0.0.0:8002",
			Sink:      "0.0.0.0:8004",
			EndPoints: []string{},
			EnableEp:  false,
			Service:   "echo",
			Namespace: "test",
		},
	},
}

var _TestPipeDefinitionEqual = []PipeDefinition{
	PipeDefinition{
		Source:    "0.0.0.0:8001",
		Sink:      "0.0.0.0:8002",
		EndPoints: []string{},
		EnableEp:  false,
		Service:   "echo",
		Namespace: "test",
	},
	PipeDefinition{
		Source:    "0.0.0.0:8001",
		Sink:      "0.0.0.0:8002",
		EndPoints: []string{},
		EnableEp:  false,
		Service:   "echo",
		Namespace: "test",
	},
}

var _TestPipeDefinitionNotEqual = []PipeDefinition{
	PipeDefinition{
		Source:    "0.0.0.0:8002",
		Sink:      "0.0.0.0:8003",
		EndPoints: []string{},
		EnableEp:  false,
		Service:   "echo",
		Namespace: "test",
	},
	PipeDefinition{
		Source:    "0.0.0.0:8002",
		Sink:      "0.0.0.0:8004",
		EndPoints: []string{},
		EnableEp:  false,
		Service:   "echo",
		Namespace: "test",
	},
}

func TestPipeDefinition(t *testing.T) {
	if !_TestPipeDefinitionEqual[0].Equal(_TestPipeDefinitionEqual[1]) {
		t.Errorf("%v %v", _TestPipeDefinitionEqual[0], _TestPipeDefinitionEqual[1])
	}
	if pipe := _TestPipeDefinitionEqual[0].Copy(_TestPipeDefinitionEqual[1]); !pipe.Equal(_TestPipeDefinitionEqual[0]) || !pipe.Equal(_TestPipeDefinitionEqual[1]) {
		t.Errorf("%v %v", _TestPipeDefinitionEqual[0], _TestPipeDefinitionEqual[1])
	}
	if _TestPipeDefinitionNotEqual[0].Equal(_TestPipeDefinitionNotEqual[1]) {
		t.Errorf("%v %v", _TestPipeDefinitionNotEqual[0], _TestPipeDefinitionNotEqual[1])
	}

	// for i, lhs := range _TestPipeDefinitionEqual {
	// 	fmt.Println(lhs.Equal(_TestPipeDefinitionNotEqual[i]))
	// 	fmt.Println(lhs, _TestPipeDefinitionNotEqual[i])
	// }
	// for i, lhs := range _TestPipeDefinitionEqual {
	// 	fmt.Println(lhs.Equal(_TestPipeDefinitionNotEqual[len(_TestPipeDefinitionEqual)-i-1]))
	// 	fmt.Println(lhs, _TestPipeDefinitionNotEqual[len(_TestPipeDefinitionEqual)-i-1])
	// }
}

func TestPipeDefinitionMap(t *testing.T) {
	m := _testPipe1
	// fmt.Println(m)
	equal := m["Equal"]
	notequal := m["!Equal"]
	if !equal[0].Equal(equal[1]) {
		t.Errorf("%v %v", equal[0], equal[1])
	}
	if notequal[0].Equal(notequal[1]) {
		t.Errorf("%v %v", notequal[0], notequal[1])
	}

	if p1, p2 := equal[0].Copy(notequal[0]), equal[1].Copy(notequal[1]); !p1.Equal(notequal[0]) || !p2.Equal(notequal[1]) {
		t.Errorf("%v %v", p1, p2)
	}
	if p1 := equal[1].Copy(notequal[1]); !p1.Equal(notequal[1]) {
		t.Errorf("%v %v", p1, notequal[1])
	}
	if p2 := equal[1].Copy(notequal[1]); !p2.Equal(notequal[1]) {
		t.Errorf("%v %v", p2, notequal[1])
	}
	if !equal[0].Equal(equal[1]) {
		t.Errorf("Copy Value Modified Reference %v %v", equal[0], equal[1])
	}
	if notequal[0].Equal(notequal[1]) {
		t.Errorf("Copy Value Modified Reference %v %v", notequal[0], notequal[1])
	}
	p1, p2 := equal[0].Copy(notequal[0]), equal[1].Copy(notequal[1])
	// fmt.Println("p1      ", p1)
	// fmt.Println("equal[0]", equal[0])
	// fmt.Println("p2      ", p2)
	// fmt.Println("equal[1]", equal[1])
	// fmt.Println(p1, p2, equal[0], equal[1])
	// fmt.Println(m)
	// fmt.Println(m)
}
