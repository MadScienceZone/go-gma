module github.com/MadScienceZone/go-gma/v5

go 1.18

//	github.com/mattn/goveralls v0.0.9 // indirect
require github.com/schwarmco/go-cartesian-product v0.0.0-20180515110546-d5ee747a6dc9

require (
	github.com/hashicorp/go-version v1.6.0
	github.com/lestrrat-go/strftime v1.0.6
	github.com/mattn/go-sqlite3 v1.14.16
	github.com/newrelic/go-agent/v3 v3.24.0
	github.com/newrelic/go-agent/v3/integrations/nrsqlite3 v1.1.1
	golang.org/x/exp v0.0.0-20221217163422-3c43f8badb15
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/grpc v1.54.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

retract v5.8.2 // missing source files
