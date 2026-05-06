# Deploy

## Server setup
Rent a VDS server. Generate SSH keys and copy your public key to the server.

> [!WARNING]
> Ensure the SSH key is successfully uploaded. If this step is skipped, SSH access to the server will be lost after the deployment pipeline secures the host.

## Reverse proxy
Although the application ensures it runs in a safe environment, it is not responsible for HTTPS and high-level routing. Therefore, without a reverse proxy, it is inaccessible from the internet.

Configure any reverse proxy. Below is an example configuration using Caddy:

```caddyfile
goroutine.mipselqq.uk {
    reverse_proxy localhost:8080
}

grafana.goroutine.mipselqq.uk {
    reverse_proxy localhost:3000
}

prometheus.goroutine.mipselqq.uk {
    reverse_proxy localhost:9090
}
```

## Configuration
Variable and secret settings are located in your repository under **Settings > Secrets and variables > Actions**.

![Variables Placeholder](placeholder)

> [!NOTE]
> Any variable or secret can be set repository-wide for both environments, or specified per environment (e.g., `staging` or `production`).

### Variables
- `DEPLOY_HOST`: Target server hostname or IP (`goroutine.mipselqq.uk`)
- `DEPLOY_USERNAME`: SSH user for deployment (`deployer`)
- `ADMIN_PORT`: Port for administrative endpoints (`9091`)
- `ALLOWED_ORIGINS`: Allowed CORS origins, comma separated, add * to list to allow any origin (`https://goroutine.mipselqq.uk`)
- `ENV`: Runtime environment (`production` or `staging`)
- `HOST`: Server interface to bind the app (`0.0.0.0`)
- `JWT_EXP`: Token expiration duration (`24h`)
- `LOG_LEVEL`: Application log level (`info`)
- `PORT`: Main application port (`8080`)
- `POSTGRES_DB`: Database name (`todo_db`)
- `POSTGRES_HOST`: Database host (`db`)
- `POSTGRES_PORT`: Database port (`5432`)
- `POSTGRES_USER`: Database user (`user`)
- `PROMETHEUS_USER`: Prometheus dashboard username (`admin`)
- `SWAGGER_HOST`: API documentation host (`goroutine.mipselqq.uk`)

### Secrets
- `DEPLOY_SSH_KEY`: Private SSH key for server access
- `GF_SECURITY_ADMIN_PASSWORD`: Admin password for Grafana
- `JWT_SECRET`: Secret key used for JWT signing
- `POSTGRES_PASSWORD`: Database password
- `PROMETHEUS_BCRYPT_HASH`: Bcrypt hash of the Prometheus password
- `PROMETHEUS_PASSWORD`: Plain password for Prometheus

## Continuous deployment
After the deploy action is triggered, the application will be automatically built, transferred, and run on the server.
