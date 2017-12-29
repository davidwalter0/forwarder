package impl

import (
	"context"
	"log"
	// "os"

	pb "github.com/davidwalter0/forwarder/rpc/pipe"
	empty "github.com/golang/protobuf/ptypes/empty"
	// log "github.com/davidwalter0/logwriter"
)

// var writer, log = logwriter.NewLogWriter()

// // GetLogger return the implementations default logwriter pair
// func GetLogger() (*logwriter.LogWriter, *_log.Logger) {
// 	return writer, log
// }

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
	// Logger := log.New(os.Stderr, "ERROR", log.Lshortfile)
	stream, err := client.Watch(context.Background(), &empty.Empty{})
	if err != nil {
		// Logger.Printf("%v.Watch(_) = _, %v", client, err)
		log.Printf("%v.Watch(_) = _, %v", client, err)
	}
	var row uint64
	for {
		if pipe, err := stream.Recv(); err != nil {
			// Logger.Printf("%v.Recv() got error %v, want %v\n", stream, err, nil)
			log.Printf("%v.Recv() got error %v, want %v\n", stream, err, nil)
			break
		} else {
			// Logger.Println(pipe.ToString(row))
			log.Println(pipe.ToString(row))
			row++
		}
	}
}
