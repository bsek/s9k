module github.com/bsek/s9k

go 1.23

replace github.com/oslokommune/common-lib-go/aws => ../common-lib-go/aws

require (
	github.com/aws/aws-sdk-go-v2 v1.32.3
	github.com/aws/aws-sdk-go-v2/config v1.27.33
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.40.7
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.39.0
	github.com/aws/aws-sdk-go-v2/service/ecr v1.36.3
	github.com/aws/aws-sdk-go-v2/service/ecs v1.45.4
	github.com/aws/aws-sdk-go-v2/service/s3 v1.61.2
	github.com/aws/aws-sdk-go-v2/service/ssm v1.52.8
	github.com/aws/aws-sdk-go-v2/service/sts v1.30.7
	github.com/creack/pty v1.1.21
	github.com/gdamore/tcell/v2 v2.7.4
	github.com/oslokommune/common-lib-go/aws v1.3.3
	github.com/rs/zerolog v1.33.0
	github.com/samber/lo v1.39.0
	golang.org/x/oauth2 v0.19.0
	golang.org/x/term v0.24.0
	golang.org/x/text v0.18.0
)

require (
	github.com/ProtonMail/go-crypto v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.4 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.32 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.13 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.34.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.9.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.34.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.22.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.26.7 // indirect
	github.com/cloudflare/circl v1.3.7 // indirect
	github.com/gdamore/encoding v1.0.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.54.0 // indirect
	go.opentelemetry.io/otel v1.29.0 // indirect
	go.opentelemetry.io/otel/metric v1.29.0 // indirect
	go.opentelemetry.io/otel/trace v1.29.0 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/exp v0.0.0-20240404231335-c0f41cb1a7a0 // indirect
	golang.org/x/sys v0.25.0 // indirect
)

require (
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.25.8
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.22.8
	github.com/aws/aws-sdk-go-v2/service/lambda v1.58.3
	github.com/aws/smithy-go v1.22.0 // indirect
	github.com/google/go-github/v50 v50.2.0
	github.com/rivo/tview v0.0.0-20240404161134-dfc1d8680fec
)
