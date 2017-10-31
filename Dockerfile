FROM golang:1.8.3-alpine3.6 as gobuild
RUN apk update && apk add git
RUN mkdir -p /root/go
ENV GOPATH="/root/go"
RUN go get -u github.com/golang/dep/cmd/dep

RUN mkdir -p /root/go/src/main
WORKDIR /root/go/src/main
# why another user??
# RUN addgroup -S databox && adduser -S -g databox databox

ADD Gopkg.* ./
RUN $GOPATH/bin/dep ensure -vendor-only
ADD src src
RUN GGO_ENABLED=0 GOOS=linux go build -a -tags netgo -installsuffix netgo -ldflags '-d -s -w -extldflags "-static"' -o app src/app.go
ADD . .

FROM scratch
# COPY --from=gobuild /etc/passwd /etc/passwd
# USER databox
WORKDIR /root
COPY --from=gobuild /root/go/src/main/app .
COPY --from=gobuild /root/go/src/main/www/ /root/www/
COPY --from=gobuild /root/go/src/main/tmpl/ /root/tmpl/
COPY --from=gobuild /root/go/src/main/etc/ /root/etc/
COPY --from=gobuild /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
LABEL databox.type="driver"
EXPOSE 8080

CMD ["./app"]
#CMD ["sleep","2147483647"]
