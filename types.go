package main

// MappingPair name for identifying application forward: free form text
type MappingPair string

// Forward host:port pair mapping between downstream and upstream, in
// a kubernetes environment the mapping might be endpoint pairs
type Forward struct {
	DownStream string   `json:"downstream" help:"downstream ingress point host:port"`
	UpStream   string   `json:"upstream"   help:"upstream service         host:port"`
	Active     []string `json:"active"     help:"active list of connections"`
}

type Forwards map[MappingPair]Forward
