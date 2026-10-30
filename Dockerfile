FROM golang:1.22.3-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api app.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/api /api
EXPOSE 3000
CMD ["/api"]
