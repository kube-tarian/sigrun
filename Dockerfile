FROM golang:latest
# All these steps will be cached
RUN mkdir /project
WORKDIR /project
COPY go.mod .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download
# COPY the source code as the last step
COPY . .
RUN go build -o main main.go
ENTRYPOINT ./main