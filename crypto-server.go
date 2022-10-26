package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jamespearly/loggly"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func badRequestToLoggly(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	client := loggly.New("crypto-server")
	client.EchoSend("error", "Method: "+r.Method+". Not allowed from: "+r.RemoteAddr+"Path: "+r.RequestURI)
}

func sendToLoggly(r *http.Request) {
	client := loggly.New("crypto-server")
	client.EchoSend("info", "Source ip: "+r.RemoteAddr+". Path: "+r.RequestURI+". Method: "+r.Method)
}

type ServerConf struct {
	Host string `yaml:"server-host"`
	Port string `yaml:"server-port"`
}

func handleGetStatus(w http.ResponseWriter, r *http.Request) {
	sendToLoggly(r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(time.Now().String()))
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	serverConf := &ServerConf{
		Host: os.Getenv("HOST"),
		Port: os.Getenv("PORT"),
	}

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/maldonado/status", handleGetStatus).Methods("GET")
	r.HandleFunc("/maldonado/status", badRequestToLoggly).Methods("POST", "PUT", "DELETE", "PATCH")

	srv := &http.Server{
		Addr: ":" + serverConf.Port,

		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go func() {
		fmt.Println("Starting server...")
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	srv.Shutdown(ctx)

	log.Println("shutting down")
	os.Exit(0)

}
