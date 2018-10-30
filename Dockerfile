FROM golang:1.11.1 AS build-env

# Add namespace here to resolve /vendor dependencies
ENV NAMESPACE github.com/schjan/tlc59711
WORKDIR /build/

ADD . ./

ARG version=dev

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -v -ldflags "-w -s"  -a -installsuffix cgo -o /out/test cmd/tlc59711test/main.go

FROM scratch
ENV DOCKER=true
COPY --from=build-env /out/test /
ENTRYPOINT [ "./test" ]