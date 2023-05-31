# Use the offical Golang image to create a build artifact.
FROM golang:alpine
LABEL authors="Alex Sysoev"

# Copy local code to the container image.
COPY  go.mod go.sum /home/go/github.com/WatherMG/topmetrics/
COPY /cmd/server /home/go/github.com/WatherMG/topmetrics/cmd/server
COPY /pkg/metric /home/go/github.com/WatherMG/topmetrics/pkg/metric
COPY /logs /home/go/github.com/WatherMG/topmetrics/logs
WORKDIR /home/go/github.com/WatherMG/topmetrics


# Build the command inside the container.
RUN go build -o server ./cmd/server/main.go


## Use a Docker multi-stage build to create a lean production image.
#FROM gcr.io/distroless/base
#COPY --from=builder /go/app/ .

CMD ["./server"]