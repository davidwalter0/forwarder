```
generate from .proto file

protoc --proto_path=../pipe --proto_path=/go/src --go_out=../pipe ../pipe/pipe.proto
```

option to force generation via go build

without the plugins=grpc:write/to/path client service info not generated
contrast
- protoc --proto_path=../pipe --proto_path=/go/src --go_out=../pipe ../pipe/pipe.proto
vs
- protoc --proto_path=../pipe --proto_path=/go/src --go_out=plugins=grpc:../pipe ../pipe/pipe.proto

```
//go:generate protoc --proto_path=../pipe --proto_path=/go/src --go_out=plugins=grpc:../pipe ../pipe/pipe.proto
```
compile:
`protoc --proto_path=pipe --proto_path=/go/src --go_out=plugins=grpc:pipe pipe/pipe.proto; go run server/server.go `




w/o tls

```
go run client/client.go
go run server/server.go
```



with tls

```
go run client/client.go  --tls --cert-file certs/example.com.crt --key-file certs/example.com.key -server-addr example.com:10000
go run server/server.go  --tls --cert-file certs/example.com.crt --root-ca certs/RootCA.crt --key-file certs/example.com.key  -server-addr example.com:10000
```


```
protoc --proto_path=src --go_out=build/gen src/foo.proto src/bar/baz.proto
```


```
mkdir gen
protoc --proto_path=src --go_out=gen src/pipes.proto src/pipe.proto
protoc --proto_path=src --c_out=gen src/pipes.proto src/pipe.proto
# c++ version
protoc --proto_path=src --cpp_out=gen src/pipes.proto src/pipe.proto
```

```
message Pipe {
	Name      string;
	Source    string;
	Sink      string;
	repeated Endpoints EP   ;
	EnableEp  bool  ;
	Service   string;
	Namespace string;
	Debug     bool  ;
}


message Pipes {
  repeated Pipe pipe = 1;
}
```

```
// PipeDefinition maps source to sink
type PipeDefinition struct {
	Name      string `json:"name"      help:"map key"`
	Source    string `json:"source"    help:"source ingress point host:port"`
	Sink      string `json:"sink"      help:"sink service point   host:port"`
	Endpoints *EP    `json:"endpoints" help:"endpoints (sinks) k8s api / config"`
	EnableEp  bool   `json:"enable-ep" help:"enable endpoints from service"`
	Service   string `json:"service"   help:"service name"`
	Namespace string `json:"namespace" help:"service namespace"`
	Debug     bool   `json:"debug"     help:"enable debug for this pipe"`
}

```

```
// https://github.com/google/protobuf/blob/master/src/google/protobuf/empty.proto

// A generic empty message that you can re-use to avoid defining duplicated
// empty messages in your APIs. A typical example is to use it as the request
// or the response type of an API method. For instance:
//
//     service Foo {
//       rpc Bar(google.protobuf.Empty) returns (google.protobuf.Empty);
//     }
//
// The JSON representation for `Empty` is empty JSON object `{}`.
message Empty {}
```
