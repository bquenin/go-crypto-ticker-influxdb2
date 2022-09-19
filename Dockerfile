FROM golang:1.19.2-alpine3.16 as builder
WORKDIR /work

# Download module in a separate layer to allow caching for the build
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd ./cmd

RUN CGO_ENABLED=0 go build -o ticker ./cmd/ticker/main.go

FROM alpine:3.16.2
WORKDIR /bin
COPY --from=builder /work/ticker /bin/ticker
CMD /bin/ticker
