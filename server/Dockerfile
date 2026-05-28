FROM golang:1.22-alpine AS build
WORKDIR /src
COPY shared/ /src/shared/
COPY server/ /src/server/
WORKDIR /src/server
RUN CGO_ENABLED=0 go build -o /out/whitagotchi-server ./cmd/whitagotchi-server

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/whitagotchi-server /whitagotchi-server
EXPOSE 8080
ENTRYPOINT ["/whitagotchi-server"]
