FROM golang:1.19-alpine as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

# Copy the go source
COPY *.go ./

# Build go binary
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64go build -o /urlstats-webservice
RUN go build -o /sortedurlstats

FROM alpine

WORKDIR /

COPY --from=builder /sortedurlstats /sortedurlstats

# ONLY NEEDED FOR OFFLINE TESTING: jsonDataSource = GETJSONDATA_FILE
# COPY resources/raw-json-files/ resources/raw-json-files/

ENV DATA_COLLECTION_METHOD=http

EXPOSE 5000

ENTRYPOINT ["/sortedurlstats"]