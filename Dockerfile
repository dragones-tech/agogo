# --- build: binario estático, sin cgo ---
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /jehosogo .
RUN mkdir -p /data

# --- runtime: solo el binario, sin SO ---
FROM scratch
COPY --from=build /jehosogo /jehosogo
COPY --from=build --chown=65534:65534 /data /data
ENV JEHOSOGO_ADDR=:8080 \
    JEHOSOGO_DB=/data/jehosogo.db \
    JEHOSOGO_BASE_URL=http://localhost:8080
EXPOSE 8080
USER 65534:65534
ENTRYPOINT ["/jehosogo"]
