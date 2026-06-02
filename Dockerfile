# --- build: binario estático, sin cgo ---
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /agogo .
RUN mkdir -p /data

# --- runtime: solo el binario, sin SO ---
FROM scratch
COPY --from=build /agogo /agogo
COPY --from=build --chown=65534:65534 /data /data
ENV AGOGO_ADDR=:8888 \
    AGOGO_DB=/data/agogo.db \
    AGOGO_BASE_URL=http://localhost:8888
EXPOSE 8888
USER 65534:65534
ENTRYPOINT ["/agogo"]
