module github.com/thebrokencube/files-with-a-dot/cmd/dendrik

go 1.25.0

require github.com/thebrokencube/files-with-a-dot/pkg/dendrik v0.0.0

require (
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/term v0.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/thebrokencube/files-with-a-dot/pkg/dendrik => ../../pkg/dendrik
