FROM golang:1.11-alpine3.8 AS build
WORKDIR /go/src/github.com/ntrrg/tiler
COPY main.go .
RUN go install

FROM alpine3.8 as debug
COPY --from=build /go/bin /bin

FROM scratch
COPY --from=build /go/bin /bin
WORKDIR /data
VOLUME /data
EXPOSE 4000
ENTRYPOINT ["/bin/tiler"]

