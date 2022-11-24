package swagger

//go:generate rm -rf server
//go:generate mkdir -p server
//go:generate oapi-codegen -generate types -o server/fileupload-api-types.gen.go -package server swagger.yml
//go:generate oapi-codegen -generate chi-server -o server/fileupload-api-server.gen.go -package server swagger.yml
