FROM golang:1.21.4 as builder

ARG CGO_ENABLED=0
WORKDIR /support_line

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o /cmd/app/main . /support_line.go

FROM scratch
ENV TZ="Europe/Moscow"
WORKDIR /teorema
COPY --from=builder /support_line/cmd/app /support_line/cmd/app
CMD [". /support_line"]