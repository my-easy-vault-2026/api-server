module github.com/my-easy-vault-2026/api-server

go 1.24.0

replace github.com/polevpn/elog => github.com/paper-trade-chatbot/elog v1.1.8

replace shared-modules => ./

require (
	github.com/gin-gonic/gin v1.11.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/swaggo/swag v1.16.6
	golang.org/x/image v0.27.0 // indirect
)

require (
	github.com/BurntSushi/toml v1.5.0
	github.com/fogleman/gg v1.3.0 // indirect
	github.com/go-playground/validator/v10 v10.30.1
	github.com/gobeam/stringy v0.0.7
	github.com/jinzhu/copier v0.4.0
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/nicksnyder/go-i18n/v2 v2.4.0
	github.com/redis/go-redis/v9 v9.6.1
	github.com/shopspring/decimal v1.4.0
	golang.org/x/text v0.34.0
	gorm.io/driver/mysql v1.5.7
	gorm.io/gorm v1.30.0
)

require (
	github.com/getsentry/sentry-go v0.42.0
	github.com/getsentry/sentry-go/gin v0.42.0
	github.com/gin-contrib/cors v1.7.6
	github.com/gorilla/websocket v1.5.3
	github.com/joho/godotenv v1.5.1
	github.com/my-easy-vault-2026/shared-modules v0.0.0-20260212083834-eb97fd74f447
	github.com/spf13/cobra v1.10.2
	github.com/spf13/viper v1.19.0
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.1
	go.uber.org/fx v1.24.0
	go.uber.org/zap v1.26.0
	gorm.io/plugin/dbresolver v1.6.2
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/bytedance/gopkg v0.1.3 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-openapi/swag/conv v0.25.4 // indirect
	github.com/go-openapi/swag/jsonname v0.25.4 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.4 // indirect
	github.com/go-openapi/swag/loading v0.25.4 // indirect
	github.com/go-openapi/swag/stringutils v0.25.4 // indirect
	github.com/go-openapi/swag/typeutils v0.25.4 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.4 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/magiconair/properties v1.8.9 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.59.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.6.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/urfave/cli/v2 v2.27.7 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/mock v0.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.33.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/bytedance/sonic v1.15.0 // indirect
	github.com/bytedance/sonic/loader v0.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/go-openapi/jsonpointer v0.22.4 // indirect
	github.com/go-openapi/jsonreference v0.21.4 // indirect
	github.com/go-openapi/spec v0.22.3 // indirect
	github.com/go-openapi/swag v0.25.4 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mailru/easyjson v0.9.1 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.1 // indirect
	golang.org/x/arch v0.24.0 // indirect
	golang.org/x/crypto v0.48.0
	golang.org/x/exp v0.0.0-20241217172543-b2144cdd0a67 // indirect
	golang.org/x/net v0.50.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/tools v0.42.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
