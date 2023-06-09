# Use the offical Golang image to create a build artifact.
FROM golang:alpine
LABEL authors="Alex Sysoev"

# Copy local code to the container image.
COPY  go.mod go.sum /home/go/github.com/WatherMG/topmetrics/
COPY /cmd/server /home/go/github.com/WatherMG/topmetrics/cmd/server
COPY /pkg/server /home/go/github.com/WatherMG/topmetrics/pkg/server
COPY /pkg/logstash /home/go/github.com/WatherMG/topmetrics/pkg/logstash
COPY /pkg/connector /home/go/github.com/WatherMG/topmetrics/pkg/connector
COPY /pkg/metric /home/go/github.com/WatherMG/topmetrics/pkg/metric
COPY /logs /home/go/github.com/WatherMG/topmetrics/logs
WORKDIR /home/go/github.com/WatherMG/topmetrics


# Build the command inside the container.
RUN go build -o server /home/go/github.com/WatherMG/topmetrics/cmd/server/main.go


## Use a Docker multi-stage build to create a lean production image.
#FROM gcr.io/distroless/base
#COPY --from=builder /go/app/ .

CMD ["./server"]