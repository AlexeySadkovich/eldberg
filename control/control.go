package control

import (
	"context"
	"github.com/AlexeySadkovich/eldberg/node"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/AlexeySadkovich/eldberg/config"
)

type Control struct {
	node node.Service
	ws   *WSServer
	http *HTTPServer

	logger *zap.SugaredLogger
}

type ControlParams struct {
	fx.In

	Node   node.Service
	Config config.Config

	Logger *zap.SugaredLogger
}

func New(lc fx.Lifecycle, p ControlParams) *Control {
	cfg := p.Config.GetNodeConfig().Control
	c := &Control{
		node:   p.Node,
		logger: p.Logger,
	}

	for _, v := range cfg {
		switch v.Protocol {
		case "ws":
			c.ws = newWSServer(v.Port)
		case "http":
			c.http = newHTTPServer(v.Port)
		default:
			c.logger.Infof("unknown protocol [%s], skipped", v.Protocol)
		}
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go c.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			c.Shutdown(ctx)
			return nil
		},
	})

	return c
}

// Start runs all required servers
func (c *Control) Start() {
	if c.ws != nil {
		c.logger.Info("starting ws server...")
		go func() {
			if err := c.ws.run(); err != nil {
				c.logger.Errorf("start ws server failed: %s", err)
			}
		}()
	}

	if c.http != nil {
		c.logger.Info("starting http server...")
		go func() {
			if err := c.http.run(); err != nil {
				c.logger.Errorf("start http server failed: %s", err)
			}
		}()
	}

	select {}
}

func (c *Control) Shutdown(ctx context.Context) {
	if c.ws != nil {
		if err := c.ws.shutdown(ctx); err != nil {
			c.logger.Errorf("shutdown ws server: %s", err)
		}
	}

	if c.http != nil {
		if err := c.http.shutdown(ctx); err != nil {
			c.logger.Errorf("shutdown http server: %s", err)
		}
	}
}
