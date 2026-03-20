package main

import (
	"log"
	"net/http"

	"momentum-api/internal/db"
	"momentum-api/internal/routes"
)

func main() {
	database := db.InitDB()
	defer database.Close()

	mux := routes.RegisterRoutes(database)
	handler := enableCORS(mux)

	log.Println("Server running on http://localhost:8080")

	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal(err)
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
