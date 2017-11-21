FROM golang
MAINTAINER Alsey <zamber@gmail.com>

RUN go get github.com/gorilla/mux
RUN go get github.com/go-sql-driver/mysql
RUN go get gopkg.in/redis.v5

ADD . /go/src/uid-gen

RUN go install uid-gen

ENTRYPOINT [ "/go/bin/uid-gen" ]

EXPOSE 3000
