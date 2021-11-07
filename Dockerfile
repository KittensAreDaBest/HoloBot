FROM golang:1.17-alpine AS build
WORKDIR /app
COPY go.mod go.sum /app
RUN go mod download
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w"
    -v \
    -trimpath \
    -o holobot \
    main.go
RUN upx holobot

FROM gcr.io/distroless/static-debian11
COPY --from=build /app/holobot /
CMD ["/holobot"]
