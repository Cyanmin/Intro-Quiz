# Build stage
FROM golang:1.21 AS build
# Build a static binary to avoid GLIBC version mismatch
ENV CGO_ENABLED=0
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go .
RUN go build -o server

# Run stage
FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=build /app/server ./
EXPOSE 8080
CMD ["./server"]
