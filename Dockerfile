FROM golang:1.26.9-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /shop ./cmd/shop

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /shop /shop
COPY --from=build /src/spec/fixtures/demo/katalog.json /demo/katalog.json
ENTRYPOINT ["/shop"]
