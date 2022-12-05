FROM golang:1.19-alpine as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

# Copy the go source
COPY *.go ./

# Build go binary
RUN go build -o /sortedurlstats

FROM alpine

WORKDIR /

COPY --from=builder /sortedurlstats /sortedurlstats


ENV DATA_COLLECTION_METHOD=http

EXPOSE 5000

ENTRYPOINT ["/sortedurlstats"]