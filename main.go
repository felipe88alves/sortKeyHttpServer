package main

import (
	"log"
	"os"
)

const (
	envVarUrlSource = "DATA_COLLECTION_METHOD"
	envVarUrlPath   = "DATA_COLLECTION_PATH"
)

func main() {
	dataSourceType := os.Getenv(envVarUrlSource)
	dataSourcePath := os.Getenv(envVarUrlPath)

	svc, err := newUrlStatDataService(dataSourceType, dataSourcePath)
	if err != nil {
		panic(err)
	}
	svc = newLoggingService(svc)

	apiServer := newApiServer(svc)
	log.Fatal(apiServer.start(":5000"))
}
