FROM golang:1.22 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/lispflow ./cmd/api

FROM gcr.io/distroless/static
COPY --from=builder /app/lispflow /app/lispflow
COPY schema.sql /app/schema.sql
EXPOSE 8080
CMD ["/app/lispflow"]
