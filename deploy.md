# Deploy to Fly.io

This guide explains how to deploy the StoneWeigh application to Fly.io.

## Prerequisites

- [Fly.io account](https://fly.io/)
- [flyctl CLI installed](https://fly.io/docs/getting-started/installing-flyctl/) (optional, you can use the web dashboard)
- Docker installed locally (for testing)
- GitHub repository (for web-based deployment)

## Steps

### 1. Install flyctl (if not already installed)

```bash
curl -L https://fly.io/install.sh | sh
```

### 2. Authenticate with Fly.io

```bash
fly auth login
```

This will open a browser window for you to log in to your Fly.io account.

### 3. Prepare for Go Build

Since you want to use `go build` with SQLite and without GoCV, ensure there's no `Dockerfile` in the project root (rename or delete it if it exists). This will allow Fly.io to use its built-in Go buildpack.

If you have a `Dockerfile`, run:

```bash
mv Dockerfile Dockerfile.backup
```

### 4. Initialize the Fly app

From the project root directory:

```bash
fly launch
```

This command will:
- Ask you to choose an app name (e.g., stoneweigh)
- Select a region for deployment (choose one close to your users)
- Generate a `fly.toml` configuration file
- Use the Go buildpack to build your application (since no Dockerfile is present)

### 5. Configure fly.toml (if needed)

The `fly launch` command creates a `fly.toml` file. For this Go application using the buildpack, it should look something like this:

```toml
app = "stoneweigh"
primary_region = "sin"

[build]
  # Using Go buildpack - no dockerfile specified

[env]
  GIN_MODE = "release"
  PORT = "8080"

[[services]]
  internal_port = 8080
  protocol = "tcp"

  [services.concurrency]
    hard_limit = 25
    soft_limit = 20

  [[services.ports]]
    handlers = ["http"]
    port = "80"

  [[services.ports]]
    handlers = ["tls", "http"]
    port = "443"
```

The Go buildpack will automatically run `go build` without any special tags, ensuring SQLite is used and GoCV is not activated. Adjust other settings based on your application requirements. Make sure the PORT matches your application's configuration.

### 5. Set Environment Variables

Before deploying, set your environment variables:

```bash
fly secrets set SESSION_SECRET="your-secure-random-string"
fly secrets set ADMIN_USERNAME="admin"
fly secrets set ADMIN_PASSWORD="strong-password"
fly secrets set DB_DRIVER="sqlite"
fly secrets set DB_DSN="stoneweigh.db"
```

For database, if using PostgreSQL, you can attach a Fly.io Postgres:

```bash
fly postgres create
fly postgres attach <postgres-app-name>
```

### 7. Deploy the application

```bash
fly deploy
```

This will build your Go application using the buildpack and deploy it to Fly.io. The first deployment may take several minutes.

### 7. Check deployment status

```bash
fly status
```

To view logs:

```bash
fly logs
```

### 8. Access your application

Once deployed, Fly.io will provide a URL for your application (e.g., https://stoneweigh.fly.dev). You can also check it with:

```bash
fly open
```

## Alternative: Build and Deploy via Fly.io Web Dashboard

If you prefer not to use the CLI, you can build and deploy through the Fly.io web interface using GitHub integration:

### 1. Connect Your Repository

1. Go to [Fly.io Dashboard](https://fly.io/dashboard)
2. Click "Launch an app"
3. Choose "Connect to GitHub" instead of "From a Dockerfile"
4. Authorize Fly.io to access your GitHub account
5. Select your StoneWeigh repository

### 2. Configure the App

- Choose an app name and region
- Fly.io will detect it's a Go application and set up the buildpack automatically
- Review the generated configuration (similar to fly.toml)

### 3. Set Environment Variables

In the dashboard:
- Go to your app's settings
- Under "Secrets", add the same environment variables as in step 5 above:
  - SESSION_SECRET
  - ADMIN_USERNAME
  - ADMIN_PASSWORD
  - DB_DRIVER=sqlite
  - DB_DSN=stoneweigh.db

### 4. Deploy

- Push your code to GitHub (if not already)
- In the dashboard, click "Deploy" or set up auto-deploy on push
- The web interface will build your Go application using the buildpack and deploy it

### 5. Monitor and Access

- View build logs and app status in the dashboard
- Access your app URL from the dashboard
- Scale, view metrics, and manage the app through the web interface

This method is convenient if you want to avoid installing flyctl locally.

## Additional Commands

- **Update deployment**: Run `fly deploy` again after making changes to the code
- **Scale the app**: `fly scale count 2` (for 2 instances)
- **View environment**: `fly secrets list`
- **SSH into the app**: `fly ssh console`

## Database Considerations

- For SQLite (default): Data persists in the Fly volume. Make sure to configure a volume if needed.
- For PostgreSQL: Use Fly's managed Postgres for better performance and reliability.

## Troubleshooting

- If deployment fails, check the logs with `fly logs`
- Ensure your `go.mod` file is properly configured and all dependencies are listed
- The Go buildpack will automatically handle building with `go build` without GoCV tags, using SQLite as the database
- For hardware integration (serial ports, RTSP), note that Fly.io runs in containers, so physical hardware access may require different setup
- Make sure the application listens on the PORT environment variable (default 8080)

For more detailed documentation, visit [Fly.io Docs](https://fly.io/docs/).