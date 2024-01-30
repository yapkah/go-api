FROM golang:latest

ENV GOPROXY https://goproxy.cn,direct
WORKDIR $GOPATH/src/github.com/smartblock/gta-api
COPY . $GOPATH/src/github.com/smartblock/gta-api
RUN go build .

EXPOSE 8000
ENTRYPOINT ["./go-gin-example"]
