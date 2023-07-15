FROM golang:1.20.6-alpine3.18 as builder
WORKDIR /work

# Download module in a separate layer to allow caching for the build
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd ./cmd

RUN CGO_ENABLED=0 go build -o ticker ./cmd/ticker/main.go

FROM alpine:3.18.2
WORKDIR /bin
COPY --from=builder /work/ticker /bin/ticker
CMD /bin/ticker
