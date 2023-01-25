package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	delegatedidentityv1 "github.com/spiffe/spire-api-sdk/proto/spire/api/agent/delegatedidentity/v1"
	"github.com/spiffe/spire-api-sdk/proto/spire/api/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO:
//
// - log the details of the agent SVID once it's got (should be able to get from the client) Nope, we can't :(


func main() {

	var socketPath string

	flag.StringVar(&socketPath, "s", "/run/spire/sockets/admin.sock", "The path for the agent Unix socket.")
	flag.Parse()

	selectors := flag.Args()

	if len(selectors) == 0 {
		log.Fatal("Must include at least one selector")
		}
		
	for _, selector := range selectors {
		colonCount := strings.Count(selector, ":")
		if colonCount == 0 {
			log.Fatalf("Selector %s is incorrectly formatted, needs at least one colon\n", selector)
		}
	}

	firstTime := true

	if _, err := os.Stat(socketPath); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("%s does not exist\n", socketPath)
	}

	for {
		if !firstTime {
			// TODO: use backoff?
			time.Sleep(10 * time.Second)
		} else {
			firstTime = false
		}

		stream, err := InitWatcher(socketPath, selectors...)

		if err != nil {
			log.Printf("%s\n", err)
			continue
		}

		for {
			
			resp, err := stream.Recv()

			if err != nil {
				log.Printf("Spiffe: error fetching X509-SVID %v\n", err)
				stream.CloseSend()
				break
			}

			log.Printf("Spiffe: handling X509-SVID rotation - %s\n", time.Now().String())
			for _, svid := range resp.X509Svids {
				log.Printf("Spiffe: processing spiffe://%s%s, Expires at %s\n", 
							svid.X509Svid.Id.TrustDomain,
							svid.X509Svid.Id.Path,
							time.Unix(svid.X509Svid.ExpiresAt, 0))
			}

			

		}
	}
}

// InitWatcher connects to spire control plane
// Cilium can subscribe to spire based on pod selectors and start receiving
// SVID updates.
func InitWatcher(sockPath string, selectbase ...string) (delegatedidentityv1.DelegatedIdentity_SubscribeToX509SVIDsClient, error) {

	unixPath := "unix://" + sockPath

	conn, err := grpc.Dial(unixPath, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("spiffe: grpc.Dial() failed on %s: %s", sockPath, err)
	}

	client := delegatedidentityv1.NewDelegatedIdentityClient(conn)

	var selectors []*types.Selector

	for _, s := range selectbase {
		items := strings.Split(s, ":")
		typekey := items[0]
		value := strings.Join(items[1:], ":")

		selectors = append(selectors, &types.Selector{
			Type: typekey,
			Value: value,
		})
	}

	fmt.Printf("Selectors\n:%#v\n", selectors)

	req := &delegatedidentityv1.SubscribeToX509SVIDsRequest{
		Selectors: selectors,
	}

	stream, err := client.SubscribeToX509SVIDs(context.Background(), req)

	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("spiffe: stream failed on %s: %s", sockPath, err)
	}

	return stream, nil
}

