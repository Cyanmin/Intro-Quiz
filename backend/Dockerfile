# Build stage
FROM golang:1.21 AS build
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go .
RUN go build -o server

# Run stage
FROM gcr.io/distroless/base-debian11
WORKDIR /app
COPY --from=build /app/server ./
EXPOSE 8080
CMD ["./server"]
