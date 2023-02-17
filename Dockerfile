FROM golang:1.18.10 AS builder

ENV GOPROXY      https://goproxy.io

RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go mod tidy
RUN go build -o  chatgpt-dingtalk  -buildvcs=false


FROM centos:centos7
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/ .
RUN chmod +x chatgpt-dingtalk && yum -y install vim net-tools telnet wget curl && yum clean all

RUN rm config.*
CMD ./chatgpt-dingtalk
