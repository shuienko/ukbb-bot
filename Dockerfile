FROM golang:1.16
LABEL maintainer="oleksandr.shuienko@gmail.com"

WORKDIR /go/src/ukbb-bot
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["ukbb-bot"]