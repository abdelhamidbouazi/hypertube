FROM golang:1.25-bookworm

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./

RUN go mod graph | awk '{if ($1 !~ "@") print $2}' | xargs go get

COPY . .

# RUN go build -o /usr/local/bin/hypertube-serve ./main.go

# CMD ["hypertube-serve"]
CMD ["air"]
