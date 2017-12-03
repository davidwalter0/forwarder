package pipe

import (
	"fmt"
	"time"

	ts "github.com/golang/protobuf/ptypes/timestamp"
)

// string          Name              = 3 ;
// string          Source            = 4 ;
// string          Sink              = 5 ;
// bool            EnableEp          = 6 ;
// string          Service           = 7 ;
// string          Namespace         = 8 ;
// bool            Debug             = 9 ;
// repeated string Endpoints         = 10;
// Mode            Mode              = 11;
// google.protobuf.Timestamp When    = 12 ;
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

// ToString from a pipe
func (l *PipeLog) ToString(row uint64) string {
	// return fmt.Sprintf("\n%-15s%-15s%-15s%-9.9s%-15s%-15s%-15s%-15s%-15s%-32s\n",
	// 	"Name", "Source", "Sink", "EnableEp", "Service", "Namespace", "Debug", "Endpoints", "Mode", "When") +
	// return fmt.Sprintf("\n%-20.20s%-15s%-15s%-15s%-9.9s%-15s%-15s%-15s%-15s\n",
	// 	"When", "Name", "Source", "Sink", "EnableEp", "Service", "Namespace", "Mode", "Endpoints") +
	// 	fmt.Sprintf("%-20.20v%-15.15v%-15.15v%-15.15v%-9.9v%-15.15v%-15.15v%-15.15v%-15.15v\n",
	// 		// fmt.Sprintf("%-15.15v%-15.15v%-15.15v%-9.9v%-15.15v%-15.15v%-15.15v%-15.15v%-15.15v%-32v\n",
	if row%MAXROWS == 0 {
		return fmt.Sprintf("%-20.20v%-15v%-15v%-15.15v%-15v%-15v%-5v%-15s\n",
			"When", "Name", "Source", "Sink", "Service", "Namespace", "Mode", "Endpoints") +
			// fmt.Sprintf("%-20.20v%-15.15v%-15.15v%-15.15v%-9.9v%-15.15v%-15.15v%-15.15v%-15.15v\n",
			fmt.Sprintf("%-20.20v%-15.15v%-15.15v%-15.15v%-15.15v%-15.15v%-5.5v%-v",
				// fmt.Sprintf("%-15.15v%-15.15v%-15.15v%-9.9v%-15.15v%-15.15v%-15.15v%-15.15v%-15.15v%-32v\n",
				ToTime(l.PipeInfo.When).String()[:19],
				l.PipeInfo.Name,
				l.PipeInfo.Source,
				l.PipeInfo.Sink,
				// l.PipeInfo.EnableEp,
				l.PipeInfo.Service,
				l.PipeInfo.Namespace,
				// l.PipeInfo.Debug,
				l.PipeInfo.Mode,
				l.PipeInfo.Endpoints)
	} else {
		return fmt.Sprintf("%-20.20v%-15.15v%-15.15v%-15.15v%-15.15v%-15.15v%-5.5v%-v",
			// fmt.Sprintf("%-15.15v%-15.15v%-15.15v%-9.9v%-15.15v%-15.15v%-15.15v%-15.15v%-15.15v%-32v\n",
			ToTime(l.PipeInfo.When).String()[:19],
			l.PipeInfo.Name,
			l.PipeInfo.Source,
			l.PipeInfo.Sink,
			// l.PipeInfo.EnableEp,
			l.PipeInfo.Service,
			l.PipeInfo.Namespace,
			// l.PipeInfo.Debug,
			l.PipeInfo.Mode,
			l.PipeInfo.Endpoints)
	}
}
