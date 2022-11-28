package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
)

type FileUploadServer struct {

}

func NewFileUploadServer() *FileUploadServer{
	return &FileUploadServer{}
}

// This function wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendError(w http.ResponseWriter, code int, message string) {
	serverError := Error{
		Code:    int32(code),
		Message: message,
	}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(serverError)
}

func (f *FileUploadServer) CheckHealth(w http.ResponseWriter, r *http.Request) {
	var database HealthStatusDatabase = HealthStatusDatabaseOK
	healthz := HealthStatus {
		Server: HealthStatusServerOK,
		Database: &database,
	}

	h, _ := json.Marshal(healthz)
	w.Write(h)
} 

func (f *FileUploadServer) UploadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(5 << 20) //5mb
	response := []ApiResponse{}

	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid file.")
		return
	}

	for _, h := range r.MultipartForm.File["file"] {
		var status ApiResponseStatus = SUCCESS
		if err := saveFile(h); err != nil {
			log.Printf("Filename: %v - %v", h.Filename, err)
			status = FAIL
		}
	
		resp := ApiResponse {
			Filename: &h.Filename,
			Status: &status,
			Date: &openapi_types.Date{
				Time: time.Now().UTC(),
			},
		}
		response = append(response, resp)
	}

	v, err := json.Marshal(response)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Error occurred while generating api response.")
	} 
	
	w.Write(v)

}

func saveFile(fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("Error reading file. %v", err)
	}
	defer file.Close()

	f, err := os.OpenFile("./uploads/" + fileHeader.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("Error creating file. %v", err)
	}
	defer f.Close()

	_, err = io.Copy(f, file)

	if err != nil {
		return fmt.Errorf("Error saving the file. %v", err)
	}

	return nil
}

