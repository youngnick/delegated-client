// sclient connects to the SPIRE agent and watches whatever SPIFFE IDs it gets
// issued.
package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"time"

	"github.com/spiffe/go-spiffe/workload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)


func main() {

	var socketPath string

	flag.StringVar(&socketPath, "s", "/run/spire/sockets/agent/agent.sock", "The path for the agent Unix socket.")
	flag.Parse()

	if _, err := os.Stat(socketPath); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("%s does not exist\n", socketPath)
	}

	unixPath := "unix://" + socketPath

	watcher := &SVIDWatcher{}
	client, err := workload.NewX509SVIDClient(watcher, workload.WithAddr(unixPath), workload.WithGRPCOptions(grpc.WithTransportCredentials(insecure.NewCredentials())))
	if err != nil {
		log.Fatalf("Couldn't initialize a SPIFFE client: %s", err)
	}

	err = client.Start()
	if err != nil {
		log.Fatalf("Couldn't start the client: %s", err)
	}

	for {
		// wait.
		time.Sleep(1000)
	}	


}

type SVIDWatcher struct {
	SVIDs *workload.X509SVIDs
}

func (sw *SVIDWatcher) UpdateX509SVIDs(svids *workload.X509SVIDs) {
	sw.SVIDs = svids

	defaultSVID := sw.SVIDs.Default()
	log.Printf("Received a new SVID:\n%#v\n", defaultSVID)
}

func (sw *SVIDWatcher) OnError(err error) {
	log.Fatalf("%s\n", err)
}
