module github.com/phantom-proxy/phantom-proxy/v14

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
)

require (
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	go.uber.org/mock v0.4.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/exp v0.0.0-20241210194714-1829a127f884 // indirect
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/tools v0.29.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)
