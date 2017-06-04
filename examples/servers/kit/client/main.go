package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/NYTimes/gizmo/examples/servers/kit/api"
)

var (
	serverAddr = flag.String("server_addr", "127.0.0.1:8081", "The server address in the format of host:port")

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
			log.Print("unable to close gRPC connection: ", err)
		}
	}()

	client := api.NewApiServiceClient(conn)

	if *mostPopularFlag {
		mostPop, err := client.GetMostPopularResourceTypeSectionTimeframe(context.Background(),
			&api.GetMostPopularResourceTypeSectionTimeframeRequest{
				ResourceType: "mostviewed",
				Section:      "all-sections",
				Timeframe:    int32(7),
			},
		)
		if err != nil {
			log.Fatal("get most popular list error: ", err)
		}

		fmt.Println("Most Popular Results:")
		out, _ := json.MarshalIndent(mostPop, "", "    ")
		fmt.Fprint(os.Stdout, string(out))
		fmt.Println("")
	}

	if *catsFlag {
		cats, err := client.GetCats(context.Background(), &google_protobuf.Empty{})
		if err != nil {
			log.Fatal("get cats list: ", err)
		}

		fmt.Println("Most Recent Articles on 'Cats':")
		out, _ := json.MarshalIndent(cats, "", "    ")
		fmt.Fprint(os.Stdout, string(out))
	}
}
