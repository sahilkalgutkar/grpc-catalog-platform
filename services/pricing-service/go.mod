module github.com/skalgutkar/grpc-catalog-platform/services/pricing-service

go 1.25.0

require (
	github.com/skalgutkar/grpc-catalog-platform/gen v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.9.0
	google.golang.org/grpc v1.83.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260803160001-6ac0973c030d // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/skalgutkar/grpc-catalog-platform/gen => ../../gen
