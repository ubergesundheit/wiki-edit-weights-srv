FROM golang:1-buster

WORKDIR /app

COPY . .

RUN go mod download

RUN GOOS=linux CGO_ENABLED=0 go build -o wsclient -tags netgo -ldflags "-s -w -extldflags -static" -a *.go

FROM gcr.io/distroless/static:nonroot

COPY --from=0 /app/wsclient /wsclient

ENTRYPOINT [ "/wsclient" ]
