package main

import (
	"net"
)

// MappingPair name for identifying application forward: free form text
type MappingPair string

// Forward host:port pair mapping between downstream and upstream, in
// a kubernetes environment the mapping might be endpoint pairs
type Forward struct {
	DownStream string   `json:"downstream" help:"downstream ingress point host:port"`
	UpStream   string   `json:"upstream"   help:"upstream service         host:port"`
	Active     []string `json:"active"     help:"active list of connections"`
}

// Forwards loaded from a yaml map formatted with a connection name
// and a pair of host:port strings downstream and upstream
type Forwards map[MappingPair]Forward

// Cfg options to configure forwarder
type Cfg struct {
	File  string `json:"file" doc:"yaml format file to import mappings from\n        name:\n          downstream: host:port\n          upstream:   host:port\n        " default:"/var/lib/forwarder/forwards.yaml"`
	Debug bool   `json:"debug" doc:"increase verboseness"`
}

// Forwarder a connection initiated by the return from listen and the
// up/down stream host:port pairs
type Forwarder struct {
	Connection net.Conn
	ID         uint64
	UpStream   string
	DownStream string
}

// Build info text
var Build string

// Commit git string
var Commit string

var config Cfg
var forwards Forwards
