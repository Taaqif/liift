# Stage 1: Build frontend
FROM node:22-alpine AS build-web
WORKDIR /app
COPY web/package.json web/package-lock.json ./
RUN npm ci --ignore-scripts
COPY web/ .
RUN npm run build

# Stage 2: Build Go binary
# CGO is required for the SQLite driver (mattn/go-sqlite3)
FROM golang:1.25.4-alpine AS build-go
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=build-web /app/dist ./web/dist

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags "-s -w \
      -X liift/api/handlers.Version=${VERSION} \
      -X liift/api/handlers.Commit=${COMMIT} \
      -X 'liift/api/handlers.BuildTime=${BUILD_TIME}'" \
    -o ./bin/liift \
    .

# Stage 3: Minimal runtime image
FROM alpine:3.21

# ca-certificates: for TLS; tzdata: for timezone support; wget: for HEALTHCHECK
RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S liift \
    && adduser -S liift -G liift

WORKDIR /app
COPY --from=build-go /app/bin/liift ./liift

RUN mkdir -p /data /storage/images \
    && chown -R liift:liift /app /data /storage/images

USER liift

# /data: SQLite database file; /storage/images: uploaded images
VOLUME ["/data", "/storage/images"]

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -qO- http://localhost:3000/api/system || exit 1

LABEL org.opencontainers.image.title="Liift" \
      org.opencontainers.image.description="Self-hosted workout tracking application" \
      org.opencontainers.image.version="${VERSION}"

CMD ["./liift"]
