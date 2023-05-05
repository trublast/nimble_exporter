FROM golang:1.19.7 as build
ADD .  /build
RUN cd /build &&  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /nimble_exporter && strip /nimble_exporter

FROM scratch as app
COPY --from=build /nimble_exporter /nimble_exporter
CMD ["/nimble_exporter"]
