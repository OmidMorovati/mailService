package main

import (
	"database/sql"
	"github.com/alexflint/go-arg"
	"log"
	"mailingService/grpcapi"
	"mailingService/jsonapi"
	"mailingService/mdb"
	"sync"
)

var args struct {
	DbPath   string `arg:"env:MAILING_SERVICE_DB"`
	BindPath string `arg:"env:MAILING_SERVICE_BIND_JSON"`
	BindGrpc string `arg:"env:MAILING_SERVICE_BIND_GRPC"`
}

func main() {
	arg.MustParse(&args)

	if args.DbPath == "" {
		args.DbPath = "./mail.db"
	}

	if args.BindPath == "" {
		args.BindPath = ":8080"
	}

	if args.BindGrpc == "" {
		args.BindGrpc = ":8082"
	}

	log.Printf("Using db at %s", args.DbPath)
	log.Printf("Starting server at %s", args.BindPath)

	db, err := sql.Open("sqlite3", args.DbPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	mdb.TryCreate(db)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		log.Println("Starting JSON email server ...")
		jsonapi.Serve(db, args.BindPath)
		defer wg.Done()
	}()

	wg.Add(1)
	go func() {
		log.Println("Starting gRPC email server ...")
		grpcapi.Serve(db, args.BindGrpc)
		defer wg.Done()
	}()
	wg.Wait()
}
