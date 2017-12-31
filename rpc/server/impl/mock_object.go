package impl

import (
	pb "github.com/davidwalter0/forwarder/rpc/pipe"
)

// MockPipeInfo test object
func MockPipeInfo() *pb.PipeInfo {
	return &pb.PipeInfo{
		Key:       "Example",
		Source:    "0.0.0.0:80",
		Sink:      "10.30.0.80",
		EnableEp:  false,
		Namespace: "",
		Name:      "",
		Debug:     false,
		Endpoints: []string{},
		Mode:      pb.Mode_P2P,
	}
}

// MockPipeGen create an example pipe
func MockPipeGen(which string) *pb.PipeLog {
	return &pb.PipeLog{
		Timestamp: Now(),
		Text:      which,
		PipeInfo: &pb.PipeInfo{
			Key:       "Example",
			Source:    "0.0.0.0:80",
			Sink:      "10.30.0.80",
			EnableEp:  false,
			Namespace: "",
			Name:      "",
			Debug:     false,
			Endpoints: []string{},
			Mode:      pb.Mode_P2P,
		},
	}
}
