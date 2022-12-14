// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.3 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
)

// Defines values for ApiResponseStatus.
const (
	FAIL    ApiResponseStatus = "FAIL"
	SUCCESS ApiResponseStatus = "SUCCESS"
)

// Defines values for HealthStatusDatabase.
const (
	HealthStatusDatabaseOK HealthStatusDatabase = "OK"
)

// Defines values for HealthStatusServer.
const (
	HealthStatusServerOK HealthStatusServer = "OK"
)

// ApiResponse defines model for ApiResponse.
type ApiResponse struct {
	Date     *openapi_types.Date `json:"date,omitempty"`
	Filename *string             `json:"filename,omitempty"`
	Status   *ApiResponseStatus  `json:"status,omitempty"`
}

// ApiResponseStatus defines model for ApiResponse.Status.
type ApiResponseStatus string

// Error defines model for Error.
type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// HealthStatus defines model for HealthStatus.
type HealthStatus struct {
	Database *HealthStatusDatabase `json:"database,omitempty"`
	Server   HealthStatusServer    `json:"server"`
}

// HealthStatusDatabase defines model for HealthStatus.Database.
type HealthStatusDatabase string

// HealthStatusServer defines model for HealthStatus.Server.
type HealthStatusServer string

// UploadFileMultipartBody defines parameters for UploadFile.
type UploadFileMultipartBody struct {
	File openapi_types.File `json:"file"`
}

// UploadFileMultipartRequestBody defines body for UploadFile for multipart/form-data ContentType.
type UploadFileMultipartRequestBody UploadFileMultipartBody

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get server status
	// (GET /healthz)
	CheckHealth(w http.ResponseWriter, r *http.Request)
	// uploads a file
	// (POST /uploadFile)
	UploadFile(w http.ResponseWriter, r *http.Request)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// CheckHealth operation middleware
func (siw *ServerInterfaceWrapper) CheckHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.CheckHealth(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// UploadFile operation middleware
func (siw *ServerInterfaceWrapper) UploadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.UploadFile(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshallingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshallingParamError) Error() string {
	return fmt.Sprintf("Error unmarshalling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshallingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{})
}

type GorillaServerOptions struct {
	BaseURL          string
	BaseRouter       *mux.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r *mux.Router) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r *mux.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options GorillaServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = mux.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.HandleFunc(options.BaseURL+"/healthz", wrapper.CheckHealth).Methods("GET")

	r.HandleFunc(options.BaseURL+"/uploadFile", wrapper.UploadFile).Methods("POST")

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/9RVz3PrNBD+VzwLx9RKW04+UUofZHhDGTLvxOSwkdexHrIkpHXS0Mn/zkhykzgJ0wDD",
	"4d1kaX98u/vt51eQtnPWkOEA1SsE2VKH6fjg1K8UnDWB4qfz1pFnRemxRk63jfUdMlT5YgK8dQQVBPbK",
	"rGA3gUZpMtiNjYfnC+aBkfuUgUzfQfUbzD89Pj7N5zCBDw+zj7CYvB9mt7+xy88kOQZ+8t768zqkrcfQ",
	"lOH7u0NIZZhW5GOEjkLAVbI+z+fpj155qiPiFPNgv7iA5kdCze18X+xZc3GJue1vbXj+6arSJxDIr8n/",
	"C9eTKoY45+ijIb0weYP6eyszYgrSK8fKGqjggzJ1YXsuOuupwGU8zje4io2cQO81VNAyu0qIkK9LZSN2",
	"ZRqbp2IYJaciOlTRHp1iwu7bscNpXk290xbrYkPLWICScRJaSRpYnKkIDw5lS8VdOT0FtNlsSkyvpfUr",
	"MbgG8XH2+PTz/OnmrpyWLXc6ZmfyXXhu5kOiC0WJZCJisxXraDL0oTjCOvgXN8WzI/Pwy6y4T7jW5EMu",
	"7Laclre3MaV1ZNApqOC+nJb3MAGH3KYZiDaR6s94XhGfj+UH4iJPtRjWLMXzGN9nNVTw2JL8PXMTIhvy",
	"9qfod9Pp22TIpODonFYyOYvPIWZ4k494+tpTAxV8JQ76IgZxESP2Jz6NgYZeSgqh6XVh/YAvT7vBXvM/",
	"wqGYuvAeoCwOB9lA73F7CVlv6MWRZKoLyj5x4/quQ7/9mw4zrkLcpzwdWEQPkScfOZC234ZUyXgYnw42",
	"eTMp8He23p6U3/WalUPPIm74TdSOcQfG2tIMOfdysFQmgn9PDpLfJTHY/UeiXDWg43/RFWM6JpA7ItA3",
	"/we2q8kzM2vUqi5SK78EOmeWhgIz5AOXm714RT7vfzrx7fVITkMlhLYSdWsDC3RKrG9ht9j9FQAA///D",
	"X9BhdQgAAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
