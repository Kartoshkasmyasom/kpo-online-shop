package app

import (
	"context"
	"log"
	"net/http"

	"orders/internal/db"
	"orders/internal/httpapi"
	"orders/internal/kafka"
	"orders/internal/store"
)

func Run() {
	db, err := db.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	st := store.NewOrdersStore(db)

	ctx := context.Background()

	prod, err := kafka.NewSyncProducer()
	if err != nil {
		log.Fatal(err)
	}

	pub := kafka.NewOutboxPublisher(db, prod)
	go pub.Run(ctx)

	resCons := kafka.NewPaymentResultConsumer(db)
	go resCons.Run(ctx)

	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, st)

	log.Println("orders listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
