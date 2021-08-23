package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"go.uber.org/fx"

	"github.com/AlexeySadkovich/eldberg/config"
	"github.com/AlexeySadkovich/eldberg/internal/blockchain"
	"github.com/AlexeySadkovich/eldberg/internal/holder"
	"github.com/AlexeySadkovich/eldberg/internal/network"
	"github.com/AlexeySadkovich/eldberg/internal/node/service"
	"github.com/AlexeySadkovich/eldberg/internal/rpc/server"
	"github.com/AlexeySadkovich/eldberg/internal/rpc/server/controller"
	"github.com/AlexeySadkovich/eldberg/internal/storage"
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
		log.Fatalf("start node: %w", err)
	}

	<-app.Done()

	if err := app.Stop(ctx); err != nil {
		log.Fatal("stop node: %w", err)
	}
}
