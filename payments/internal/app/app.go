package app

import (
	"context"
	"log"
	"net/http"

	"payments/internal/db"
	"payments/internal/httpapi"
	"payments/internal/kafka"
	"payments/internal/store"
)

func Run() {
	db, err := db.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	st := store.NewStore(db)
	ctx := context.Background()
	cons := kafka.NewPaymentRequestConsumer(db, st)
	go cons.Run(ctx)

	prod, err := kafka.NewSyncProducer()
	if err != nil {
		log.Fatal(err)
	}
	pub := kafka.NewOutboxPublisher(db, prod)
	go pub.Run(ctx)


	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, st)

	log.Println("payments listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
