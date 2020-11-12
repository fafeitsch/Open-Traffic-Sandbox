package main

import (
	"fmt"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/bus"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/osrmclient"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/server"
	"log"
	"net/http"
)

func main() {
	fmt.Printf("Loading scenario file …\n")
	mdl, err := model.Init("samples/wuerzburg(fictional)")
	if err != nil {
		log.Fatalf("could not understand scenario file: %v", err)
	}
	fmt.Printf("Scenario loaded successfully, here is some information about it:\n")
	fmt.Println()
	fmt.Printf("%v", mdl)
	fmt.Println()
	fmt.Printf("Starting simulation …")

	clientContainer := server.NewClientContainer()
	http.Handle("/sockets", clientContainer)
	http.Handle("/", http.FileServer(http.Dir("webfrontend/dist/webfrontend")))

	publisher := func(position model.BusPosition) {
		clientContainer.BroadcastJson(position)
	}
	dispatcher := bus.NewDispatcher(mdl, publisher, osrmclient.NewRouteService("http://localhost:5000/"))
	go dispatcher.Run(mdl.Start())
	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
