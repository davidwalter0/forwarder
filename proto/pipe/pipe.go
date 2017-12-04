package pipe

import (
	"fmt"
	"time"

	ts "github.com/golang/protobuf/ptypes/timestamp"
)

var rows uint64

// Timestamp local
type Timestamp ts.Timestamp

// Time converted from Timestamp
func (ts *Timestamp) ToTime() time.Time {
	return time.Unix(ts.Seconds, int64(ts.Nanos))
}

// Time converted from Timestamp
func ToTime(ts *ts.Timestamp) time.Time {
	return time.Unix(ts.Seconds, int64(ts.Nanos))
}

var MAXROWS uint64 = uint64(15)

// ToString from a PipeLog
func (l *PipeLog) ToString(row uint64) string {
	if row%MAXROWS == 0 {
		return fmt.Sprintf("%-20.20v%-15v%-15v%-15.15v%-15v%-15v%-5v%-15s\n",
			"Timestamp", "Name", "Source", "Sink", "Service", "Namespace", "Mode", "Endpoints") +
			fmt.Sprintf("%-20.20v%-15.15v%-15.15v%-15.15v%-15.15v%-15.15v%-5.5v%-v",
				ToTime(l.PipeInfo.Timestamp).String()[:19],
				l.PipeInfo.Name,
				l.PipeInfo.Source,
				l.PipeInfo.Sink,
				l.PipeInfo.Service,
				l.PipeInfo.Namespace,
				l.PipeInfo.Mode,
				l.PipeInfo.Endpoints)
	} else {
		return fmt.Sprintf("%-20.20v%-15.15v%-15.15v%-15.15v%-15.15v%-15.15v%-5.5v%-v",
			ToTime(l.PipeInfo.Timestamp).String()[:19],
			l.PipeInfo.Name,
			l.PipeInfo.Source,
			l.PipeInfo.Sink,
			l.PipeInfo.Service,
			l.PipeInfo.Namespace,
			l.PipeInfo.Mode,
			l.PipeInfo.Endpoints)
	}
}
