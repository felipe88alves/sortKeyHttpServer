package main

import (
	"log"
	"os"
)

const (
	EnvVarUrlSource = "DATA_COLLECTION_METHOD"
	EnvVarUrlPath   = "DATA_COLLECTION_PATH"
)

func main() {
	dataSourceType := os.Getenv(EnvVarUrlSource)
	dataSourcePath := os.Getenv(EnvVarUrlPath)

	svc := newUrlStatDataService(dataSourceType, dataSourcePath)
	svc = NewLoggingService(svc)

	apiServer := NewApiServer(svc)
	log.Fatal(apiServer.Start(":5000"))
}
