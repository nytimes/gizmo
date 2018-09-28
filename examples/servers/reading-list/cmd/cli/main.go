package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"

	readinglist "github.com/nytimes/gizmo/examples/servers/reading-list"
)

var (
	host     = flag.String("host", "localhost:8081", "the host of the reading list server")
	insecure = flag.Bool("insecure", false, "use an insecure connection")

	mode = flag.String("mode", "list", "(list|update)")

	// list
	limit = flag.Int("limit", 20, "limit for the number of links to return when listing links")

	// update
	article = flag.String("url", "", "the URL to add or delete")
	delete  = flag.Bool("delete", false, "delete this URL from the list (requires -mode update)")

	creds  = flag.String("creds", "", "the path of the service account credentials file. if empty, uses Google Application Default Credentials.")
	fakeID = flag.String("fakeID", "", "for local development - a user ID to inject into the request")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	var copts []grpc.CallOption
	if *creds != "" {
		cs, err := oauth.NewJWTAccessFromFile(*creds)
		if err != nil {
			exitf("%v", err)
		}
		copts = append(copts, grpc.PerRPCCredentials(cs))
	}

	// add a fake user to the metadata for local dev
	if *fakeID != "" {
		b, err := json.Marshal(map[string]string{"id": *fakeID})
		if err != nil {
			exitf("%v", err)
		}
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(
			map[string]string{
				"x-endpoint-api-userinfo": base64.StdEncoding.EncodeToString(b),
			}))
	}

	var dopts []grpc.DialOption
	if *insecure {
		dopts = append(dopts, grpc.WithInsecure())
	} else {
		cs, err := oauth.NewApplicationDefault(ctx)
		if err != nil {
			exitf("%v", err)
		}
		copts = append(copts, grpc.PerRPCCredentials(cs))

	}

	conn, err := grpc.Dial(*host, dopts...)
	if err != nil {
		exitf("%v", err)
	}

	c := readinglist.NewReadingListServiceClient(conn)

	switch *mode {
	case "list":
		l, err := c.GetListLimit(ctx, &readinglist.GetListLimitRequest{Limit: int32(*limit)},
			copts...)
		if err != nil {
			exitf("unable to get links: %s", err.Error())
		}
		fmt.Printf("successful request with %d links returned\n", len(l.Links))
		for _, lk := range l.Links {
			fmt.Println("* " + lk.Url)
		}
	case "update":
		aurl := *article
		if len(aurl) == 0 {
			exitf("missing -url flag")
		}
		fmt.Println("saving URL:", aurl)
		m, err := c.PutLink(ctx, &readinglist.PutLinkRequest{
			Request: &readinglist.LinkRequest{
				Link:   &readinglist.Link{Url: aurl},
				Delete: *delete}},
			copts...)
		if err != nil {
			exitf("unable to update link: %v", err)
		}
		fmt.Println(m.Message)
	default:
		fmt.Println("INVALID MODE. Please choose 'update' or 'list'")
	}
}

func exitf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
	fmt.Fprintln(os.Stderr)
	os.Exit(2)
}
