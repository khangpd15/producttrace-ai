package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/app"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/config"
)

type HealthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := HealthResponse{
		Service: "go-core-service",
		Status:  "ok",
	}

	json.NewEncoder(w).Encode(response)
}

func main() {

	database, err := config.ConnectPostgres()

	if err != nil {
		log.Fatalf("database connect failed: %v", err)
	}

	dbGORM, err := config.ConnectGORM()
	if err != nil {
		panic("failed to connect database")
	}

	// rabbitURL := os.Getenv("RABBITMQ_URL")

	// conn, ch, err := rabbitmq.Connect(rabbitURL)
	// if err != nil {
	// 	log.Fatalf("rabbitmq connect failed: %v", err)
	// }

	// defer conn.Close()
	// defer ch.Close()

	// if err := rabbitmq.SetupTopology(ch); err != nil {
	// 	log.Fatalf("rabbitmq topology failed: %v", err)
	// }

	// log.Println("RabbitMQ connected")
	// log.Println("RabbitMQ topology initialized")
	// pub := publisher.New(ch, rabbitmq.EventExchange)

	// port := os.Getenv("PORT")
	// if port == "" {
	// 	port = "8080"
	// }

	appli := app.NewApp(database, dbGORM)

	if err := appli.Router.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v \n", err)
	}

	// mux := http.NewServeMux()
	// mux.HandleFunc("/health", healthHandler)
	// mux.HandleFunc("/test-event", func(w http.ResponseWriter, r *http.Request) {

	// 	event := types.Event{
	// 		EventID:       uuid.NewString(),
	// 		EventType:     rabbitmq.ProductCreatedRK,
	// 		EventVersion:  "1.0",
	// 		Timestamp:     time.Now().UTC(),
	// 		Producer:      "go-core-service",
	// 		CorrelationID: uuid.NewString(),
	// 		Payload: map[string]any{
	// 			"productId": "p-001",
	// 			"name":      "Coffee Bean",
	// 		},
	// 	}

	// 	if err := pub.Publish(event); err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}

	// 	w.WriteHeader(http.StatusOK)
	// 	_, _ = w.Write([]byte("event published"))
	// })

	// log.Printf("Go service is running on port %s\n", port)

	// if err := http.ListenAndServe(":"+port, mux); err != nil {
	// 	log.Fatal(err)
	// }
}
