# Personal media streamer

ui and go microservices to stream available media viles

# MVP

## Backend

- list all media in a lib
- stream selected file
- URl to trigger yt-dl script
- when script finished downloading send notification
- download endpoint

## UI

- see all available media
- play video audio files
- background play
- download file from existing
- download from URL

## POST MVP

- auth for multiple users
- user lib / home screen
- subscription (run cron to allow update)
- playlists
- sort played / new files
- allow local files to display and stream

## Services/downloader — Build & Run

Build the container (run from the repository root so the local replace directive for `mymedia/common` resolves):

```bash
docker build -f Services/downloader/Dockerfile -t mymedia/downloader:latest .
```

Run the container (exposes the Go microservice on port 8080 and mounts the host media library into the container):

```bash
docker run --rm -p 8080:8080 -v /ABS/PATH/TO/media_library:/media_library mymedia/downloader:latest
```

Notes:

- The Dockerfile used is [`Services/downloader/Dockerfile`](Services/downloader/Dockerfile:1).
- The built Go binary includes `mymedia/common` (e.g. `mymedia/common/cors`) because the image is built from the repository root so the `replace` directive in [`Services/downloader/go.mod`](Services/downloader/go.mod:1) resolves.
- The service writes downloads into `/media_library` inside the container which should be mounted to your host `media_library` directory.
