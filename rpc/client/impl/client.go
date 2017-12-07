package impl

import (
	"context"
	"fmt"
	"log"

	pb "github.com/davidwalter0/forwarder/rpc/pipe"
	empty "github.com/golang/protobuf/ptypes/empty"
)

// RunPipeInfoRequest connects to pipeinfo service and prints the configuration
func RunPipeInfoRequest(client pb.WatcherClient, pipeName *pb.PipeName) {
	if pipeInfo, err := client.GetPipe(context.Background(), pipeName); err != nil {
		log.Println(err)
	} else {
		log.Println(*pipeInfo)
	}
}

// RunPipeLogClient connects to pipe log service and monitors the logs
func RunPipeLogClient(client pb.WatcherClient) {
	stream, err := client.Watch(context.Background(), &empty.Empty{})
	if err != nil {
		log.Fatalf("%v.Watch(_) = _, %v", client, err)
	}
	var row uint64
	for {
		if pipe, err := stream.Recv(); err != nil {
			log.Printf("%v.Recv() got error %v, want %v\n", stream, err, nil)
			break
		} else {
			fmt.Println(pipe.ToString(row))
			row++
		}
	}
}
