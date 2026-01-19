module github.com/ashwinyue/next-show

go 1.24.0

require (
	github.com/PuerkitoBio/goquery v1.11.0
	github.com/cloudwego/eino v0.7.21
	github.com/cloudwego/eino-ext/callbacks/cozeloop v0.0.0-00010101000000-000000000000
	github.com/cloudwego/eino-ext/components/embedding/dashscope v0.0.0-00010101000000-000000000000
	github.com/cloudwego/eino-ext/components/embedding/openai v0.0.0-00010101000000-000000000000
	github.com/cloudwego/eino-ext/components/model/ark v0.0.0-00010101000000-000000000000
	github.com/cloudwego/eino-ext/components/model/openai v0.0.0-00010101000000-000000000000
	github.com/cloudwego/eino-ext/components/tool/duckduckgo v0.0.0-00010101000000-000000000000
	github.com/cloudwego/eino-ext/components/tool/mcp v0.0.0-00010101000000-000000000000
	github.com/coze-dev/cozeloop-go v0.1.20
	github.com/gin-gonic/gin v1.10.0
	github.com/google/uuid v1.6.0
	github.com/mark3labs/mcp-go v0.43.2
	github.com/spf13/viper v1.19.0
	gorm.io/driver/postgres v1.5.9
	gorm.io/gorm v1.25.12
)

replace (
	github.com/cloudwego/eino => ../
	github.com/cloudwego/eino-ext/callbacks/cozeloop => ../eino-ext/callbacks/cozeloop
	github.com/cloudwego/eino-ext/components/embedding/dashscope => ../eino-ext/components/embedding/dashscope
	github.com/cloudwego/eino-ext/components/embedding/openai => ../eino-ext/components/embedding/openai
	github.com/cloudwego/eino-ext/components/model/ark => ../eino-ext/components/model/ark
	github.com/cloudwego/eino-ext/components/model/openai => ../eino-ext/components/model/openai
	github.com/cloudwego/eino-ext/components/tool/duckduckgo => ../eino-ext/components/tool/duckduckgo
	github.com/cloudwego/eino-ext/components/tool/mcp => ../eino-ext/components/tool/mcp
)

require (
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/bluele/gcache v0.0.2 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/bytedance/gopkg v0.1.3 // indirect
	github.com/bytedance/sonic v1.14.1 // indirect
	github.com/bytedance/sonic/loader v0.3.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/cloudwego/eino-ext/libs/acl/openai v0.1.10 // indirect
	github.com/coze-dev/cozeloop-go/spec v0.1.7 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/eino-contrib/jsonschema v1.0.3 // indirect
	github.com/evanphx/json-patch v0.5.2 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.20.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/goph/emperror v0.17.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/invopop/jsonschema v0.13.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/meguminnnnnnnnn/go-openai v0.1.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nikolalohinski/gonja v1.5.3 // indirect
	github.com/nikolalohinski/gonja/v2 v2.3.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pkg/errors v0.9.2-0.20201214064552-5dd12d0cfe7f // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/slongfield/pyfmt v0.0.0-20220222012616-ea85ff4c361f // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/volcengine/volc-sdk-golang v1.0.23 // indirect
	github.com/volcengine/volcengine-go-sdk v1.1.49 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/yargevad/filepathx v1.0.0 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/arch v0.15.0 // indirect
	golang.org/x/crypto v0.44.0 // indirect
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
