FROM golang:latest

ENV GOPROXY https://goproxy.cn,direct
WORKDIR $GOPATH/src/github.com/yapkah/go-api
COPY . $GOPATH/src/github.com/yapkah/go-api
RUN go build .

EXPOSE 8000
ENTRYPOINT ["./go-gin-example"]
