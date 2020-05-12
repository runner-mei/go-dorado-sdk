// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/whywaita/go-dorado-sdk/example/lib"
)

func main() {
	ctx := context.Background()

	client, err := lib.GetClient()
	if err != nil {
		log.Fatal(err)
	}

	portalIPs, err := client.LocalDevice.GetPortalIPAddresses(ctx, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", portalIPs)

	fmt.Println("operation is done!")
}