FROM golang:1-buster

WORKDIR /app

COPY . .

RUN go mod download

RUN GOOS=linux go build -o wsclient -ldflags "-linkmode external -extldflags -static" -a *.go

FROM gcr.io/distroless/base:nonroot

COPY --from=0 /app/wsclient /wsclient

ENTRYPOINT [ "/wsclient" ]
