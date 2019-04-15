FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/jmlattanzi/itaic-backend/itaic
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

# This container exposes port 8080 to the outside world
EXPOSE 8000

# Run the executable
CMD ["itaic"]
