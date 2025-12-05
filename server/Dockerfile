FROM golang:1.25-bookworm

WORKDIR /app

# Install FFmpeg for video transcoding
RUN apt-get update && apt-get install -y ffmpeg && rm -rf /var/lib/apt/lists/*

# Create directories for volumes
RUN mkdir -p /app/downloads /app/hls_output /app/subtitles

RUN go install github.com/air-verse/air@latest

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./

RUN go mod graph | awk '{if ($1 !~ "@") print $2}' | xargs go get

COPY . .

RUN make swagger

RUN go build -o /usr/local/bin/hypertube-serve ./main.go

# CMD ["hypertube-serve"]
CMD ["sh", "-c", "make swagger && air"]
