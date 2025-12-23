package main

import(
	"net/http"
	"fmt"
	"log"
)

func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
			http.Error(w, "method is not allowed", http.StatusMethodNotAllowed)
			return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 20<<20)
	err := r.ParseMultipartForm(20 << 20)
	if err != nil {
		http.Error(w, "failed to parse form: " + err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field is required: " + err.Error(), http.StatusBadRequest)
		return
	}

	defer file.Close()
}

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})

	http.HandleFunc("/create", handleCreateOrder)

	addr := ":8081"
	log.Println("order-service starting on", addr)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("server error:", err)
	}
}