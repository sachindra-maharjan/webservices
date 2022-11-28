package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	middleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gorilla/mux"
	"github.com/webservices/fileupload-api/api"
)

func init() {
	openapi3filter.RegisterBodyDecoder("text/plain", openapi3filter.FileBodyDecoder)
	openapi3filter.RegisterBodyDecoder("text/csv", openapi3filter.FileBodyDecoder)
	openapi3filter.RegisterBodyDecoder("image/jpeg", openapi3filter.FileBodyDecoder)
}

func main() {
	var port = flag.Int("port", 8080, "Port for test HTTP server")
	flag.Parse()

	swagger, err := api.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}
	
	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	server := api.NewFileUploadServer()

	// This is how you set up a basic Gorilla router
	r := mux.NewRouter()

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	r.Use(middleware.OapiRequestValidator(swagger))

	// We now register our petStore above as the handler for the interface
	api.HandlerFromMux(server, r)

	s := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("0.0.0.0:%d", *port),
	}

	// And we serve HTTP until the world ends.
	fmt.Printf("Listening to port %v\n", *port)
	log.Fatal(s.ListenAndServe())

}