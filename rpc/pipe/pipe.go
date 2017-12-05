package pipe

import (
	"fmt"
	"time"

	"github.com/davidwalter0/forwarder/pipe"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// Now Timestamp set to current time
func Now() *timestamp.Timestamp {
	now := time.Now()
	s := now.Unix()
	n := int32(now.Nanosecond())
	return &timestamp.Timestamp{Seconds: s, Nanos: n}
}

var rows uint64

// // Timestamp local
// type Timestamp timestamp.Timestamp

// // Time converted from Timestamp
// func (ts *Timestamp) ToTime() time.Time {
// 	return time.Unix(ts.Seconds, int64(ts.Nanos))
// }

// Time converted from Timestamp
func ToTime(ts *timestamp.Timestamp) time.Time {
	return time.Unix(ts.Seconds, int64(ts.Nanos))
}

var MAXROWS uint64 = uint64(15)

// ForwardMode2RpcMode
func ForwardMode2RpcMode(mode string) (m Mode) {
	m = Mode_P2P
	switch mode {
	case "Point2Point":
		m = Mode_P2P
	case "EndPointList":
		m = Mode_EpList
	case "ServiceLookup":
		m = Mode_SvcLkup
	}
	return
}

// RpcMode2ForwardMode
func RpcMode2ForwardMode(mode Mode) (m string) {
	m = "Point2Point"
	switch mode {
	case Mode_P2P:
		m = "Point2Point"
	case Mode_EpList:
		m = "EndPointList"
	case Mode_SvcLkup:
		m = "ServiceLookup"
	}
	return
}

// Definition from a PipeLog
func (l *PipeLog) Definition() *pipe.Definition {
	return &pipe.Definition{
		Name:      l.PipeInfo.Name,
		Source:    l.PipeInfo.Source,
		Sink:      l.PipeInfo.Sink,
		Service:   l.PipeInfo.Service,
		Namespace: l.PipeInfo.Namespace,
		Mode:      RpcMode2ForwardMode(l.PipeInfo.Mode),
		Endpoints: (*pipe.EP)(&l.PipeInfo.Endpoints),
	}
}

// Pipe2PipeLog from a PipeLog
func Pipe2PipeLog(p *pipe.Definition) *PipeLog {
	var pipeInfo = &PipeInfo{
		Name:      p.Name,
		Source:    p.Source,
		Sink:      p.Sink,
		Service:   p.Service,
		Namespace: p.Namespace,
		Mode:      ForwardMode2RpcMode(p.Mode),
		Endpoints: []string(*p.Endpoints),
	}
	return &PipeLog{
		Timestamp: Now(),
		Text:      "No new log",
		PipeInfo:  pipeInfo,
	}
}

// ToString from a PipeLog
func (l *PipeLog) ToString(row uint64) string {
	if row%MAXROWS == 0 {
		return fmt.Sprintf("%-20.20v%-15v%-15v%-15.15v%-15v%-15v%-5v%-15s\n",
			"Timestamp", "Name", "Source", "Sink", "Service", "Namespace", "Mode", "Endpoints") +
			fmt.Sprintf("%-20.20v%-15.15v%-15.15v%-15.15v%-15.15v%-15.15v%-5.5v%-v",
				ToTime(l.Timestamp).String()[:19],
				l.PipeInfo.Name,
				l.PipeInfo.Source,
				l.PipeInfo.Sink,
				l.PipeInfo.Service,
				l.PipeInfo.Namespace,
				l.PipeInfo.Mode,
				l.PipeInfo.Endpoints)
	}
	return fmt.Sprintf("%-20.20v%-15.15v%-15.15v%-15.15v%-15.15v%-15.15v%-5.5v%-v",
		ToTime(l.Timestamp).String()[:19],
		l.PipeInfo.Name,
		l.PipeInfo.Source,
		l.PipeInfo.Sink,
		l.PipeInfo.Service,
		l.PipeInfo.Namespace,
		l.PipeInfo.Mode,
		l.PipeInfo.Endpoints)

}
