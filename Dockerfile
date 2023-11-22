FROM golang:1.21.4 as build
RUN mkdir -p /go/src/app
ADD main.go /go/src/app/
WORKDIR /go/src/app
RUN  go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod init nginx-reloader
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -o nginx-reloader .
# main image
FROM nginx
COPY --from=build /go/src/app/nginx-reloader /
CMD ["/nginx-reloader"]