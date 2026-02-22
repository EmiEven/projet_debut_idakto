# build => on utilise une image officielle Go
FROM golang:1.25.7 AS builder

WORKDIR /app

# pour copier les fichiers de modules & télécharger les dépendances
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# pour compiler app en binaire Linux statique / server = nom du binaire
RUN CGO_ENABLED=1 GOOS=linux go build -o server .

# image finale
FROM debian:stable-slim

WORKDIR /app

# pour copier le binaire compilé
COPY --from=builder /app/server .


# Installer les libs nécessaires à go-sqlite3
RUN apt-get update && apt-get install -y \ 
    libsqlite3-0 \ 
    ca-certificates \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/* 
    
COPY --from=builder /app/server .
# pour copier la base SQLite si elle existe
COPY users.db .

# port
EXPOSE 8080

# lancer le serveur
CMD ["./server"]
