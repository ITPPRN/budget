# Stage 1: Build Stage
FROM golang:1.25.4-alpine3.22 AS builder

# ติดตั้ง git และ build-base สำหรับ dependencies บางตัวที่อาจต้องการ C compiler
RUN apk add --no-cache git build-base

WORKDIR /app

# 1. Copy เฉพาะไฟล์จำพวก go.mod/sum เพื่อทำ Caching Layer
COPY go.mod go.sum ./
RUN go mod download

# 2. Copy source code ทั้งหมด
COPY . .

# 3. Build binary (ใช้ CGO_ENABLED=0 เพื่อให้รันบน Alpine ได้โดยไม่มีปัญหา library)
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o main -ldflags="-s -w" ./app/main.go

# Stage 2: Final Runtime Stage
FROM alpine:3.22

# ติดตั้ง timezone และ certificates
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Bangkok

WORKDIR /app

# Copy binary จาก builder
COPY --from=builder /app/main .

# .env จะ mount ผ่าน docker-compose หรือ env_file แทน ไม่ COPY เข้า image

EXPOSE 8000

ENTRYPOINT ["./main"]