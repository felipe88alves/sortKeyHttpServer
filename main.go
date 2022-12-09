package main

import (
	"log"
	"os"

	"github.com/felipe88alves/sortKeyHttpServer/api"
)

const (
	envVarUrlSource = "DATA_COLLECTION_METHOD"
	envVarUrlPath   = "DATA_COLLECTION_PATH"
)

func main() {
	dataSourceType := os.Getenv(envVarUrlSource)
	dataSourcePath := os.Getenv(envVarUrlPath)

	svc, err := api.NewUrlStatDataService(dataSourceType, dataSourcePath)
	if err != nil {
		panic(err)
	}
	svc = api.NewLoggingService(svc)

	apiServer := api.NewApiServer(svc)
	log.Fatal(apiServer.Start(":5000"))
}
