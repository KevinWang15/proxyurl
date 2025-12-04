# Proxy URL Fetcher

Small HTTP service that reads a proxy URL from `config.json`, accepts a URL-encoded destination in the `route` query string, calls the destination through the configured proxy (with optional basic auth embedded in the URL), and streams the upstream response back to the caller.

## Configuration

Create `config.json` in the project root:

```json
{
  "proxy_url": "https://user:password@proxy-host:port"
}
```

`proxy_url` must be an absolute HTTP(S) proxy URL. Basic-auth credentials in the URL are supported.

## Run locally

```bash
go run main.go
# listens on :8080 (override with PORT env var)
```

Fetch a target URL (the `route` value must be URL-encoded):

```bash
TARGET="https://example.com/data"
curl "http://localhost:8080/fetch?route=$(python3 -c 'import urllib.parse,os;print(urllib.parse.quote(os.environ[\"TARGET\"]))')"
```

## Docker
- Build image:
  ```bash
  docker build -t proxyurl .
  ```
- Run by mounting your `config.json` (not baked into the image):
  ```bash
  docker run -p 8080:8080 \
    -v $(pwd)/config.json:/app/config.json:ro \
    proxyurl
  ```
- If you mount the file elsewhere, set `CONFIG_PATH`:
  ```bash
  docker run -p 8080:8080 \
    -v $(pwd)/config.json:/config/config.json:ro \
    -e CONFIG_PATH=/config/config.json \
    proxyurl
  ```

## GitHub Actions

`.github/workflows/build.yml` runs `go vet`, `go test`, and `go build` on pushes and PRs.

On pushes to `main`/`master`, the workflow also builds and pushes a Docker image to `ghcr.io/<owner>/proxyurl` tagged with the branch name, the commit SHA, and `latest` when on the default branch. Set `packages: write` permissions on the repo if needed.
