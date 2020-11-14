package main

import (
	"context"
	"fmt"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/bus"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/osrmclient"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/server"
	"github.com/urfave/cli/v2"
	"log"
	"net/http"
	"os"
	"sync"
)

type options struct {
	bindAddress string
	otrsServer  string
	frequency   float64
	warp        float64
	busSpeedKmh int
}

func main() {
	app := cli.NewApp()
	app.Name = "Open Traffic Sandbox"
	app.Usage = "OTS"
	app.HelpName = "otsserver"
	app.HideHelp = true
	app.Version = "v.1.0.0alpha"
	app.Description = "Open Traffic Sandbox (OTS) is a project for simulating public transportation networks."
	app.Authors = []*cli.Author{{Name: "Fabian Feitsch", Email: "info@fafeitsch.de"}}

	options := options{}

	app.Flags = []cli.Flag{
		&cli.StringFlag{Name: "bindAddress", Usage: "Sets the bind address and port for the app", Value: "127.0.0.1:9551", Destination: &options.bindAddress},
		&cli.StringFlag{Name: "otrsServer", Usage: "The OTRS base URL for fetching route information", Value: "http://127.0.0.1:5000/", Destination: &options.otrsServer},
		&cli.Float64Flag{Name: "frequency", Usage: "The number of simulation cycles in one second.", Value: 1, Destination: &options.frequency},
		&cli.Float64Flag{Name: "warp", Usage: "Defines the relation between frequency and real time. warp=1 is real time, warp=2 lets time pass twice as fast.", Value: 1, Destination: &options.warp},
		&cli.IntFlag{Name: "busSpeed", Usage: "The constant speed of the busses (in kmh).", Value: 40, Destination: &options.busSpeedKmh},
	}

	app.Action = runWithOptions(&options)
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func runWithOptions(options *options) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		logger := log.New(os.Stdout, "", log.LstdFlags)
		logger.Printf("Loading scenario file …\n")
		mdl, err := model.Init("samples/wuerzburg(fictional)")
		if err != nil {
			return fmt.Errorf("could not understand scenario directory: %v", err)
		}
		logger.Printf("Scenario loaded successfully, here is some information about it:\n")
		logger.Println()
		logger.Printf("%v", mdl)
		logger.Println()
		logger.Printf("Starting simulation.")

		clientContainer := server.NewClientContainer()
		http.Handle("/sockets", clientContainer)
		http.Handle("/", http.FileServer(http.Dir("webfrontend/dist/webfrontend")))

		publisher := func(position model.BusPosition) {
			clientContainer.BroadcastJson(position)
		}
		dispatcher := bus.NewDispatcher(mdl, publisher, osrmclient.NewRouteService(options.otrsServer))
		dispatcher.Frequency = options.frequency
		dispatcher.Warp = options.warp
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			dispatcher.Run(mdl.Start())
		}()
		srv := http.Server{Addr: options.bindAddress}
		go func() {
			logger.Printf("Listening on %s … ", options.bindAddress)
			err := srv.ListenAndServe()
			if err != nil {
				log.Fatalf("could not start srv: %v", err)
			}
		}()
		wg.Wait()
		return srv.Shutdown(context.Background())
	}
}
