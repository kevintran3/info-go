FROM golang:1.23-alpine AS build

WORKDIR /app
COPY go.* ./
COPY *.go ./

RUN go mod download
RUN CGO_ENABLED=0 go build -o /go-info


FROM alpine:3.21

COPY --from=build /go-info /app/go-info
WORKDIR /app

EXPOSE ${PORT:-8080}

CMD ["./go-info"]
