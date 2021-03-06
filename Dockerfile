FROM golang:1.15 as build

WORKDIR /src
COPY go.* ./
RUN go mod download

COPY . .
ENV CGO_ENABLED 0
ARG VERSION
RUN go build -ldflags "-X main.version=${VERSION}" -o /go/bin/main 

FROM gcr.io/distroless/static:nonroot 
COPY --from=build /go/bin/main /app/
COPY index.html /app/
WORKDIR /app
ENTRYPOINT ["/app/main"]