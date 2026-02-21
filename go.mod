# PhantomProxy v14.0 - Core Dependencies

# Go 1.21+ required
go 1.21

require (
	// HTTP/3 QUIC
	github.com/quic-go/quic-go v0.44.0
	github.com/quic-go/quic-go/http3 v0.44.0

	// HTTP framework
	github.com/gofiber/fiber/v2 v2.52.0

	// TLS spoofing
	github.com/refraction-networking/utls v1.7.0

	// WebSocket
	github.com/gorilla/websocket v1.5.1

	// Database
	github.com/jackc/pgx/v5 v5.5.0
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/redis/go-redis/v9 v9.3.0

	// Event Bus
	github.com/nats-io/nats.go v1.34.0

	// Logging
	go.uber.org/zap v1.27.0

	// Config
	github.com/spf13/viper v1.18.0
	github.com/joho/godotenv v1.5.1

	// Utils
	github.com/google/uuid v1.5.0
	gopkg.in/yaml.v3 v3.0.1

	// Testing
	github.com/stretchr/testify v1.9.0

	// Playwright
	github.com/playwright-community/playwright-go v0.5200.1
)
