FROM golang:1.6.3-alpine
MAINTAINER coolcow <18686224@qq.com>
ENV GOBIN /go/bin
#COPY glide.yaml /go/glide.yaml
COPY src /go/src
COPY config /go/config
COPY vendor /go/vendor2/src
WORKDIR /go
ENV GOPATH /go:/go/vendor2
#RUN curl https://glide.sh/get | sh
#RUN glide install
RUN go build src/app.go
#RUN rm -rf pkg src glide.yaml
RUN ls
ENTRYPOINT go run src/app.go
#ENTRYPOINT ls
EXPOSE 12006
