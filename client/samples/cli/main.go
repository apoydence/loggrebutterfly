package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/apoydence/loggrebutterfly/client"
	v2 "github.com/apoydence/loggrebutterfly/pb/loggregator/v2"
	pb "github.com/apoydence/loggrebutterfly/pb/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var (
	verbose = flag.Bool("verbose", false, "Verbose mode")

	masterAddr      = flag.String("master", "", "The address for the master")
	sourceUUID      = flag.String("source-uuid", "", "The source-uuid to interact with")
	writePacketSize = flag.Uint("packetSize", 1024, "The size of each write packet")

	tail       = flag.Bool("tail", false, "The tail the source-uuid")
	writeData  = flag.Bool("write", false, "Write data from STDIN")
	listRoutes = flag.Bool("list", false, "List the routes")
)

func main() {
	flag.Parse()

	if !*verbose {
		grpclog.SetLogger(log.New(ioutil.Discard, "", log.LstdFlags))
	}

	if *masterAddr == "" {
		log.Fatal("You must provide a master address")
	}

	if *tail {
		tailCommand(client.New(*masterAddr))
		return
	}

	if *writeData {
		writeDataCommand(client.New(*masterAddr))
		return
	}

	if *listRoutes {
		list()
		return
	}

	onlyOneCommandUsage()
}

func setupMasterClient() pb.MasterClient {
	conn, err := grpc.Dial(*masterAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect to scheduler: %s", err)
	}
	return pb.NewMasterClient(conn)
}

func setupDataNodeClient(addr string) pb.DataNodeClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect to node: %s", err)
	}
	return pb.NewDataNodeClient(conn)
}

func tailCommand(client *client.Client) {
	if *writeData || *listRoutes {
		onlyOneCommandUsage()
	}

	if *sourceUUID == "" {
		log.Fatal("You must provide a source-uuid")
	}

	rx := client.ReadFrom(*sourceUUID)

	var i int
	for {
		envelope, err := rx()
		if err == io.EOF {
			time.Sleep(250 * time.Millisecond)
			continue
		}

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("(%d) %+v\n", i, envelope)
		i++
	}
}

func writeDataCommand(client *client.Client) {
	if *tail || *listRoutes {
		onlyOneCommandUsage()
	}

	defer func() {
		// Give the buffer time to clear
		time.Sleep(250 * time.Millisecond)
	}()

	var count int
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		var e v2.Envelope
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			log.Fatalf("unable to parse (via json) to v2.Envelope: %s", err)
		}

		for i := 0; i < 10; i++ {
			var err error
			if err = client.Write(&e); err == nil {
				break
			}
			log.Println("Error writing", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		count++
	}

}

func list() {
	if *writeData || *tail {
		onlyOneCommandUsage()
	}

	masterClient := setupMasterClient()

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := masterClient.Routes(ctx, new(pb.RoutesInfo))
	if err != nil {
		log.Fatal(err)
	}

	for i, r := range resp.Routes {
		fmt.Printf("Route %d: %+v\n", i, r)
	}
	fmt.Printf("Listed %d routes\n", len(resp.Routes))
}

func onlyOneCommandUsage() {
	log.Fatal("Use only one tail, write or list")
}
