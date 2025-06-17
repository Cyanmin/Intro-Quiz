# Build stage
FROM golang:1.21 AS build
# Build a static binary to avoid GLIBC version mismatch
ENV CGO_ENABLED=0
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY cmd/intro-quiz ./cmd/intro-quiz
COPY internal ./internal
COPY pkg ./pkg
COPY config ./config
RUN go build -o server ./cmd/intro-quiz

# Run stage
FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=build /app/server ./
EXPOSE 8080
CMD ["./server"]
