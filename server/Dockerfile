FROM golang:1.26.4 AS build

WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /server .

FROM gcr.io/distroless/static:nonroot
COPY --from=build /server /server
EXPOSE 6745/udp 6746/tcp
ENTRYPOINT ["/server"]
