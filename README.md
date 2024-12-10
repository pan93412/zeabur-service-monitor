# Zeabur Service Monitor

A simple service monitor that checks if a service is running. Suitable as a health check for worker services (that does not expose a health check endpoint).

## Usage

### Locally

```bash
MONITOR_SERVICE_ID="<your service id>" MONITOR_ENVIRONMENT_ID="<your environment id>" MONITOR_ZEABUR_TOKEN="<zeabur access token>" go run .
```

```bash
$ curl http://localhost:8080/alive
{"lastCheckedAt":"2024-12-11T00:07:32.982975+08:00","success":true}
```

### on Zeabur

You can deploy this service on Zeabur as a GitHub service.

Remember to fill these environment variables in the service settings:

- `MONITOR_SERVICE_ID`
- `MONITOR_ENVIRONMENT_ID`
- `MONITOR_ZEABUR_TOKEN`

## License

Apache-2.0
