package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/webservices/fileupload-api/pkg/swagger/server"
)

type FileUploadServer struct {

}

func (s FileUploadServer) CheckHealth(w http.ResponseWriter, r *http.Request) {
	var database server.HealthStatusDatabase = server.HealthStatusDatabaseOK
	healthz := server.HealthStatus {
		Server: server.HealthStatusServerOK,
		Database: &database,
	}

	h, err := json.Marshal(healthz)
	if(err != nil) {
		log.Fatal(err)
		http.Error(w, "Internal Server Error", 500)
	}

	w.Write(h)

}

func main() {
	s := FileUploadServer{}
	h := server.Handler(s)
	http.ListenAndServe(":3000", h)	
	log.Println("Listening to port 3000")
}