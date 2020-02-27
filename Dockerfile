FROM golang

WORKDIR /go/src

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o /go/bin/go-alrd

ENTRYPOINT [ "/go/bin/go-alrd" ]