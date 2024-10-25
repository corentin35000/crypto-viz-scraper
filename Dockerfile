# Étape 1 : Construction de l’application
FROM golang:1.23.2-alpine AS builder

# Installer les dépendances nécessaires
RUN apk update && apk add --no-cache git

# Définir le répertoire de travail
WORKDIR /app

# Copier les fichiers de projet et installer les dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copier le reste des fichiers de l'application
COPY . .

# Installer Air pour le rechargement automatique
RUN go install github.com/air-verse/air@latest

# Étape 2 : Exécution
CMD air -c .air.toml