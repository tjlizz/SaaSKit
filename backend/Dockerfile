FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -trimpath -ldflags="-s -w" -o /saaskit ./cmd/api

FROM alpine:3.22
RUN addgroup -S saaskit && adduser -S -G saaskit saaskit && apk add --no-cache ca-certificates tzdata
COPY --from=build /saaskit /usr/local/bin/saaskit
USER saaskit
EXPOSE 8080
ENTRYPOINT ["saaskit"]
