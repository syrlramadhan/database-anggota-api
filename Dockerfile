# Gunakan image base Go versi terbaru
FROM golang:1.24

# Buat direktori kerja
WORKDIR /app

# Salin file go.mod & go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Salin semua kode
COPY . .

# Build aplikasi
RUN go build -o main .

# Expose port backend
EXPOSE 9001

# Jalankan aplikasi dengan .env
CMD ["./main"]

