package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	analyzer "simplewebscrapper/internal"
	"time"
)

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

func main() {

	http.HandleFunc("/", handleAnalysis)

	log.Println("Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleAnalysis(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	log.Println("Analyze the submitted URL")
	result, err := analyzer.Analyze(ctx, r.FormValue("url"))
	if err != nil {
		log.Printf("Error occurred while analyzing URL: %v", err)
	}

	tmpl.Execute(w, map[string]interface{}{
		"Result": result,
		"Error":  err,
	})
}
