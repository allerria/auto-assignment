package main

import (
	"context"
	"github.com/georgiypetrov/auto-assignment/models"
	"github.com/georgiypetrov/auto-assignment/service"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.TextFormatter{ForceColors: true})

	log.Println("Starting the app...")
	db, err := models.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	srv, err := service.InitService(db, log)
	if err != nil {
		log.Fatal(err)
	}
	go srv.Serve()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt

	log.Println("Stopping app...")

	timeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	err = srv.Shutdown(timeout)
	if err != nil {
		log.WithError(err).Error("Error when shutdown app")
	}
	log.Info("The app stopped")
}
