package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/NYTimes/gizmo/examples/servers/rpc/service"
	"github.com/NYTimes/gizmo/server"
)

var (
	serverAddr = flag.String("server_addr", "127.0.0.1:8080", "The server address in the format of host:port")

	catsFlag        = flag.Bool("cats", true, "flag make the GetCats call")
	mostPopularFlag = flag.Bool("most-popular", true, "flag make the GetMostPopular call")
)

func main() {
	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			server.Log.Warn("unable to close gRPC connection: ", err)
		}
	}()

	nytClient := service.NewNYTProxyServiceClient(conn)

	if *mostPopularFlag {
		mostPop, err := nytClient.GetMostPopular(context.Background(), &service.MostPopularRequest{
			ResourceType:   "mostviewed",
			Section:        "all-sections",
			TimePeriodDays: uint32(1),
		})
		if err != nil {
			log.Fatal("get most popular list error: ", err)
		}

		fmt.Println("Most Popular Results:")
		out, _ := json.MarshalIndent(mostPop, "", "    ")
		fmt.Fprint(os.Stdout, string(out))
		fmt.Println("")
	}

	if *catsFlag {
		cats, err := nytClient.GetCats(context.Background(), &service.CatsRequest{})
		if err != nil {
			log.Fatal("get cats list: ", err)
		}

		fmt.Println("Most Recent Articles on 'Cats':")
		out, _ := json.MarshalIndent(cats, "", "    ")
		fmt.Fprint(os.Stdout, string(out))
	}
}
