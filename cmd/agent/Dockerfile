# Use the offical Golang image to create a build artifact.
FROM golang:alpine
LABEL authors="Alex Sysoev"

# Copy local code to the container image.
COPY  go.mod go.sum /home/go/github.com/WatherMG/topmetrics/
COPY /cmd/agent /home/go/github.com/WatherMG/topmetrics/cmd/agent/
COPY /pkg /home/go/github.com/WatherMG/topmetrics/pkg/
WORKDIR /home/go/github.com/WatherMG/topmetrics


# Build the command inside the container.
RUN go build -o agent ./cmd/agent/main.go



## Use a Docker multi-stage build to create a lean production image.
#FROM gcr.io/distroless/base
#COPY --from=builder /go/app/ .

CMD ["./agent", "-count", "1", "-interval", "2s", "-timeout", "10s"]
