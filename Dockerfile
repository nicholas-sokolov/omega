FROM golang:1.16

EXPOSE 8000
WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./... \
    && go install -v ./... \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

CMD ["./main"]