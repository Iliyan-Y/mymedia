# Multi-stage Dockerfile — builds the downloader service (including local ../common) and produces
# a runtime image that contains the compiled Go microservice and yt-dlp + ffmpeg.
#
# NOTE: Build this image from the repository root so the Docker build context includes the ../common module:
#   docker build -f Services/downloader/Dockerfile -t mymedia/downloader:latest .
FROM golang:1.25-alpine AS builder
WORKDIR /src

# Copy entire repository into the build image so the local replace directive
# (replace mymedia/common => ../common) resolves correctly.
COPY . .

# Build the downloader module
WORKDIR /src/Services/downloader
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/downloader .

FROM python:3.11-slim
LABEL maintainer="mymedia"

# Install ffmpeg and certificates
RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg ca-certificates && rm -rf /var/lib/apt/lists/*

# Install yt-dlp
RUN pip install --no-cache-dir yt-dlp

# Copy compiled Go binary
COPY --from=builder /out/downloader /usr/local/bin/downloader

WORKDIR /work

# Expose the microservice port
EXPOSE 8080

# Run the Go microservice as container entrypoint. The service will call yt-dlp locally
# (yt-dlp is installed in this image) and write into /media_library which should be
# mounted by the caller.
ENTRYPOINT ["/usr/local/bin/downloader"]