FROM golang:1.16.3

WORKDIR /go/src/bill-splitter
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...
RUN go build -o bill-splitter cmd/billSplitter/main.go

EXPOSE 3000

CMD ["./bill-splitter"]