# Rela

Rela is a WIP task tracker with ability to self-host it.

## Tech Stack

- **Backend**: Go (Gin framework)
- **Frontend**: Vue.js 3 + Vite
- **Database**: MongoDB
- **Web Server**: Nginx

## Prerequisites

- Docker and Docker Compose
- Go 1.24+ (for local development)
- Node.js (for local frontend development)

## Setup Instructions

### 1. Clone the Repository

```bash
git clone <repository-url>
cd rela
```

### 2. Environment Configuration

Copy the example environment file and configure it:

```bash
cp .env_example .env
```

Edit `.env` and set the following variables:

- `MONGO_CREDS`: MongoDB connection string with username and password
- `PORT`: Backend server port (default: `:8080`)
- `PEPPER`: 32-byte base64 encoded string for password hashing
- `MONGO_INITDB_ROOT_USERNAME`: MongoDB root username
- `MONGO_INITDB_ROOT_PASSWORD`: MongoDB root password

To generate a PEPPER value:
```bash
openssl rand -base64 32
```

### 3. Run with Docker Compose

Build and start all services:

```bash
docker-compose up -d
```

This will start:
- **Backend**: Available at `http://localhost:4444`
- **MongoDB**: Available at `localhost:4445`
- **Frontend** (via Nginx): Available at `http://localhost:1488`

### 4. Verify Services

Check that all containers are running:

```bash
docker-compose ps
```

You should see:
- `rela-backend`
- `mongodb`
- `rela-nginx`

## Local Development Setup

### Backend

```bash
cd backend
go mod download
go run .
```

The backend will run on the port specified in your `.env` file (default: `:8080`).

### Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend development server will run on `http://localhost:5173`.

## API Documentation

Once the backend is running, API documentation is available via Swagger at:
```
http://localhost:4444/swagger/index.html
```

## Stopping Services

```bash
docker-compose down
```

To remove volumes as well:
```bash
docker-compose down -v
```

## License

See [LICENSE](LICENSE) file for details.
