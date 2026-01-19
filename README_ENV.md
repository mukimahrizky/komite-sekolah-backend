# Environment Variables Configuration

This project uses environment variables for configuration. Create a `.env` file in the `backend` directory with the following variables:

## Setup Instructions

1. **Create `.env` file** in the `backend` directory:
   ```bash
   cd backend
   touch .env
   ```

2. **Copy the example template** and fill in your values:
   ```bash
   cp .env.example .env
   ```

3. **Edit `.env` file** with your actual values:

```env
# Database Configuration
DB_USER=root
DB_PASSWORD=your_mysql_password
DB_HOST=localhost
DB_PORT=3306
DB_NAME=komite_sekolah

# JWT Secret Key (IMPORTANT: Change this in production!)
JWT_SECRET=your-secret-key-change-in-production

# Server Configuration
SERVER_PORT=8080
```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_USER` | MySQL database username | `root` | No |
| `DB_PASSWORD` | MySQL database password | `` (empty) | No |
| `DB_HOST` | MySQL database host | `localhost` | No |
| `DB_PORT` | MySQL database port | `3306` | No |
| `DB_NAME` | MySQL database name | `komite_sekolah` | No |
| `JWT_SECRET` | Secret key for JWT token signing | `your-secret-key-change-in-production` | **Yes** (change in production!) |
| `SERVER_PORT` | Port for the HTTP server | `8080` | No |

## Security Notes

⚠️ **IMPORTANT:**
- Never commit `.env` file to version control (it's already in `.gitignore`)
- Change `JWT_SECRET` to a strong random string in production
- Use strong database passwords in production
- The `.env.example` file is safe to commit as it contains no sensitive data

## Generating a Strong JWT Secret

You can generate a secure random string for `JWT_SECRET`:

```bash
# Using OpenSSL
openssl rand -base64 32

# Or using Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"
```

## Database Setup

Before running the application, make sure MySQL is running and create the database:

```sql
CREATE DATABASE komite_sekolah CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

The application will automatically create the tables on first run.

