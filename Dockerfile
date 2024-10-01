# docker build -t registry.leafdev.top/leaf/rag-new:v0.0.1-fix .
FROM golang:latest as builder

COPY . /app

RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct && go mod download
RUN CGO_ENABLED=0 go build -ldflags "-w -s" -gcflags "-N -l" -o main .

# RUN
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main /app/main
ENTRYPOINT ["/app/main"]