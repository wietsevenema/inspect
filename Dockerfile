FROM golang:1.13 AS gobuilder

WORKDIR /app
COPY main.go main.go

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64
RUN go build main.go

FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=gobuilder /app/main /app/main

ENTRYPOINT ["./main"]