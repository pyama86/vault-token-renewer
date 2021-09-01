FROM golang:1.17 as builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o main .

FROM debian:stable-slim
WORKDIR /
COPY --from=builder /workspace/main .
RUN apt update && apt install -y ca-certificates
ENTRYPOINT ["/main"]
