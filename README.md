# API Proxy Service

A high-performance Go application that provides rate-limited API proxy functionality with support for direct access and SOCKS5/HTTP proxy routing.

## Features

- Rate limiting per API key
- Support for multiple API keys with different rate limits
- Direct access with configurable TPS (Transactions Per Second)
- SOCKS5 and HTTP proxy support
- Automatic failover between direct access and proxies
- Configurable logging levels
- CORS support
- Graceful shutdown
- Docker support

## Setup

1. Clone the repository
```bash
git clone https://github.com/pcmehrdad/simple-api-proxy.git
cd simple-api-proxy
```

2. Copy `config.example.json` to `config.json` and configure your settings
```bash
cp config.example.json config.json
```

Example configuration:
```json
{
    "domain": "https://api.example.com/",
    "PROXY_TYPE": "header",
    "KEY": "X-Api-Key",
    "VALUES": [
        {"your-api-key-1": 10},
        {"your-api-key-2": 8}
    ],
    "DIRECT_ACCESS": true,
    "DIRECT_ACCESS_TPS": 10,
    "PROXY_TPS": 5,
    "PROXIES": [
        "socks5://user:pass@proxy1.example.com:1080",
        "socks5h://user:pass@proxy2.example.com:1080"
    ]
}
```

3. Install dependencies
```bash
go mod download
```

4. Build the application
```bash
make build
```

5. Run the application
```bash
make run
```

## Docker Support

Build and run using Docker:
```bash
docker build -t api-proxy .
docker run -p 3003:3003 -v $(pwd)/config.json:/app/config.json api-proxy
```

Or using Docker Compose:
```bash
docker-compose up --build
```

## Project Structure

```
api-proxy/
├── cmd/
│   └── processor/
│       └── main.go           # Application entrypoint
├── internal/
│   ├── api/
│   │   └── client.go         # API client implementation
│   ├── config/
│   │   └── config.go         # Configuration handling
│   └── utils/
│       ├── proxy.go          # Proxy client utilities
│       ├── ratelimiter.go    # Rate limiting logic
│       └── types.go          # Common types
├── config.json               # Configuration file
├── docker-compose.yml        # Docker Compose configuration
├── Dockerfile               # Docker build instructions
└── Makefile                # Build and run commands
```

## Configuration Options

| Option | Description | Type | Required |
|--------|-------------|------|----------|
| domain | Target API domain | string | Yes |
| PROXY_TYPE | Type of proxy authentication | string | Yes |
| KEY | API key header name | string | Yes |
| VALUES | List of API keys and their rate limits | array | Yes |
| DIRECT_ACCESS | Enable direct API access | boolean | No |
| DIRECT_ACCESS_TPS | Rate limit for direct access | integer | If DIRECT_ACCESS is true |
| PROXY_TPS | Rate limit for proxy access | integer | No |
| PROXIES | List of proxy URLs | array | No |

## Usage

### Running with Different Log Levels

```bash
# Run with debug logging
LOG_LEVEL=debug make run

# Run with info logging
LOG_LEVEL=info make run

# Run with error logging only
LOG_LEVEL=error make run
```

### Making Requests

The proxy service listens on port 3003 by default. Requests are automatically distributed among available API keys and proxies while respecting rate limits.

Example request:
```bash
curl http://localhost:3003/your/api/endpoint
```

## Make Commands

- `make run`: Run the application
- `make build`: Build the binary
- `make clean`: Clean build artifacts
- `make run-debug`: Run with debug logging
- `make run-info`: Run with info logging
- `make run-error`: Run with error logging

## Future Improvements

- [ ] Add metrics and monitoring (Prometheus/Grafana)
- [ ] Implement proxy health checks
- [ ] Add support for dynamic API key management
- [ ] Add request retries with backoff
- [ ] Implement cache layer for responses
- [ ] Add comprehensive test suite
- [ ] Add support for WebSocket proxying
- [ ] Implement request/response logging
- [ ] Add admin API for runtime configuration
- [ ] Support for multiple target domains

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [logrus](https://github.com/sirupsen/logrus) for logging
- [golang/x/net](https://golang.org/x/net) for proxy support