package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/gorilla/mux"
	"github.com/jamespearly/loggly"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const DB_TABLE_NAME = "Maldonado-CryptoBro"

type Currency_t struct {
	Id                string `json:"id"`
	Rank              string `json:"rank"`
	Symbol            string `json:"symbol"`
	Name              string `json:"name"`
	Supply            string `json:"supply"`
	MaxSupply         string `json:"maxSupply"`
	MarketCapUsd      string `json:"marketCapUsd"`
	VolumeUsd24Hr     string `json:"volumeUsd24Hr"`
	PriceUsd          string `json:"priceUsd"`
	ChangePercent24Hr string `json:"changePercent24Hr"`
	Vwap24Hr          string `json:"vwap24Hr"`
}

type Status struct {
	TableName   string `json:"table"`
	RecordCount int64  `json:"recordCount"`
}

func getDBSession() *dynamodb.DynamoDB {
	return dynamodb.New(session.Must(session.NewSession(&aws.Config{
		Region:   aws.String("us-east-1"),
		Endpoint: aws.String("https://dynamodb.us-east-1.amazonaws.com"),
	})))
}

func badRequestToLoggly(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	client := loggly.New("crypto-server")
	client.EchoSend("error", "Method: "+r.Method+". Not allowed from: "+r.RemoteAddr+"Path: "+r.RequestURI)
}

func throwLogError(msg string) {
	client := loggly.New("crypto-server")
	client.EchoSend("error", msg)
}

func sendToLoggly(r *http.Request) {
	client := loggly.New("crypto-server")
	client.EchoSend("info", "Source ip: "+r.RemoteAddr+". Path: "+r.RequestURI+". Method: "+r.Method)
}

func handleGetAll(w http.ResponseWriter, r *http.Request) {
	db := getDBSession()

	var resp []Currency_t

	err := db.ScanPages(&dynamodb.ScanInput{
		TableName: aws.String(DB_TABLE_NAME),
	}, func(page *dynamodb.ScanOutput, last bool) bool {
		recs := []Currency_t{}

		err := dynamodbattribute.UnmarshalListOfMaps(page.Items, &recs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			throwLogError(err.Error())
			return false
		}

		resp = append(resp, recs...)

		return true
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		throwLogError(err.Error())
		return
	}

	itemJson, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		throwLogError(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write(itemJson)
	sendToLoggly(r)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	db := getDBSession()

	filt := expression.Name("Id").Equal(expression.Value(mux.Vars(r)["name"]))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		throwLogError(err.Error())
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(DB_TABLE_NAME),
	}

	result, err := db.Scan(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		throwLogError(err.Error())
	}

	currency := Currency_t{}

	for _, item := range result.Items {

		err = dynamodbattribute.UnmarshalMap(item, &currency)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			throwLogError(err.Error())
		}

	}

	resultJson, err := json.Marshal(currency)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		throwLogError(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultJson)

	sendToLoggly(r)
}

func handleGetStatus(w http.ResponseWriter, r *http.Request) {

	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(DB_TABLE_NAME),
	}

	db := getDBSession()
	table, err := db.DescribeTable(input)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		throwLogError(err.Error())
		return
	}

	status := &Status{
		TableName:   DB_TABLE_NAME,
		RecordCount: *table.Table.ItemCount,
	}

	statusJson, err := json.Marshal(status)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		throwLogError(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(statusJson)

	sendToLoggly(r)
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/maldonado/status", handleGetStatus).Methods("GET")
	r.HandleFunc("/maldonado/status", badRequestToLoggly).Methods("POST", "PUT", "DELETE", "PATCH")

	r.HandleFunc("/maldonado/all", handleGetAll).Methods("GET")
	r.HandleFunc("/maldonado/all", badRequestToLoggly).Methods("POST", "PUT", "DELETE", "PATCH")

	r.HandleFunc("/maldonado/search/{name:[-a-zA-Z]+}", handleSearch).Methods("GET")
	r.HandleFunc("/maldonado/search/", badRequestToLoggly).Methods("POST", "PUT", "DELETE", "PATCH")

	srv := &http.Server{
		Addr: ":" + os.Getenv("PORT"),

		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go func() {
		fmt.Println(srv.Addr)
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
