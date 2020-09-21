package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"

	"demoapi/config"
	"demoapi/database"
	apiserver "demoapi/server"
)

//
// look in config/validate.go for configuration flag/env values
//

var (
	version = "<unknown>"
)

func main() {
	// pulls in flag files, flag values, and environment variables
	conf, err := config.Parse()
	if err != nil {
		logrus.Fatalf("configuration error: %+v", err)
	}

	logrus.SetLevel(conf.LogLevel)
	if err := run(context.Background(), conf); err != nil {
		logrus.Fatalf("%+v", err)
	}
}

func run(ctx context.Context, c *config.Configs) error {
	// TODO(sam): pass through database configs
	db, err := database.Connect(c.DBURL, nil)
	if err != nil {
		return err
	}

	logrus.Infof("starting demoapi version %q", version)
	server := &http.Server{
		Addr:         c.ServerAddr,
		WriteTimeout: c.WriteTimeout,
		ReadTimeout:  c.ReadTimeout,
		IdleTimeout:  c.IdleTimeout,
		Handler:      apiserver.New(db, c),
	}

	go func() {
		logrus.Infof("waiting for connections on %s", c.ServerURL.String())
		_ = server.ListenAndServe()
		// ignoring the possible error here
	}()

	interruptWaiter := make(chan os.Signal, 1)
	signal.Notify(interruptWaiter, os.Interrupt)
	<-interruptWaiter // block until interrupt signal received

	// set timeout incase something takes forever after interrupt
	ctx, cancel := context.WithTimeout(ctx, c.GracefulShutdownTimeout)
	defer cancel()

	go func() {
		logrus.Infof("shutting down...")
		_ = server.Shutdown(ctx)
		// ignoring the possible error here
	}()

	// wait for gracefull shutdown or canceled context
	<-ctx.Done()

	logrus.Infof("shut down")
	return nil
}
