package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	apiURL string
	wsURL  string
	port   string
)

func main() {
	// Parse flags
	flag.StringVar(&apiURL, "api-url", "http://localhost:4444/api/v1", "API URL")
	flag.StringVar(&wsURL, "ws-url", "ws://localhost:4444/ws", "WebSocket URL")
	flag.StringVar(&port, "port", "3333", "Port to run the server on")
	flag.Parse()

	// Override with environment variables if set
	if envAPIURL := os.Getenv("API_URL"); envAPIURL != "" {
		apiURL = envAPIURL
	}
	if envWSURL := os.Getenv("WS_URL"); envWSURL != "" {
		wsURL = envWSURL
	}

	http.HandleFunc("/", serveTemplate)

	log.Println("Starting server on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	var fp string
	if r.URL.Path == "/" {
		fp = "login.html"
	} else {
		fp = strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
	}

	// Return a 404 if the template doesn't exist
	info, err := os.Stat(fp)
	if err != nil || info.IsDir() {
		http.Error(w, fmt.Sprintf("404 - %s page not found", fp), http.StatusNotFound)
		return
	}

	// Parse the template files
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	// Inject the API_URL and WS_URL into the template
	data := struct {
		API_URL string
		WS_URL  string
	}{
		API_URL: apiURL,
		WS_URL:  wsURL,
	}

	// Execute the template
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
	}
}
