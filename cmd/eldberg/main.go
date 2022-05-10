package main

import (
	"context"
	"github.com/AlexeySadkovich/eldberg/blockchain"
	"github.com/AlexeySadkovich/eldberg/holder"
	"github.com/AlexeySadkovich/eldberg/network"
	"github.com/AlexeySadkovich/eldberg/node/service"
	"github.com/AlexeySadkovich/eldberg/rpc/server"
	"github.com/AlexeySadkovich/eldberg/rpc/server/controller"
	"github.com/AlexeySadkovich/eldberg/storage"
	"log"
	"os"
	"os/signal"

	"go.uber.org/fx"

	"github.com/AlexeySadkovich/eldberg/config"
	"github.com/AlexeySadkovich/eldberg/control"
)

func main() {
	app := fx.New(
		fx.Provide(provideLogger),
		fx.Provide(config.New),
		fx.Provide(provideDB),
		fx.Provide(storage.New),
		fx.Provide(holder.New),
		fx.Provide(server.New),
		fx.Provide(network.New),
		fx.Provide(blockchain.New),
		fx.Provide(service.New),
		fx.Provide(controller.New),
		fx.Provide(control.New),

		fx.Invoke(server.Register),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop)

	go func() {
		<-stop
		cancel()
	}()

	if err := app.Start(ctx); err != nil {
		log.Fatalf("start node: %s", err)
	}

	<-app.Done()

	if err := app.Stop(ctx); err != nil {
		log.Fatalf("stop node: %s", err)
	}
}
