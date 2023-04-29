package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"movieexample.com/metadata/internal/controller/metadata"
	httpHandler "movieexample.com/metadata/internal/handler/http"
	"movieexample.com/metadata/internal/repository/memory"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/consul"
)

const serviceName = "metadata"

func main() {
	var port int
	// read port number from arguments
	flag.IntVar(&port, "port", 8081, "API handler port")
	flag.Parse()

	log.Printf("Starting the movie metadata service on port %d ...", port)

	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	// register service with the service registry
	err = registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port))
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			//report healthy state every second
			if err := registry.ReportHealthyStatus(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	repo := memory.New()
	ctrl := metadata.New(repo)
	h := httpHandler.New(ctrl)
	http.Handle("/metadata", http.HandlerFunc(h.GetMetadata))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}

}
