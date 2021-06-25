FROM golang:1.16.4-alpine3.13 as buildStage
WORKDIR /root
ADD . .
RUN apk add build-base
RUN go build -o app
# For later - RUN go test -c -coverpkg=../root/... -covermode=atomic ./api_tests

FROM alpine:3.13

RUN mkdir /gn
WORKDIR /gn

COPY --from=buildStage /root/app app
COPY --from=buildStage /root/js js
# Document that the service listens on the specified port.
EXPOSE 8092

# (For persistence setup an external volume mount and point `-db` there)
CMD ["/gn/app", "-svr", "-db", "/gn/go_notes.sqlite"]
