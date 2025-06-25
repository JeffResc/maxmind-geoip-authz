module github.com/jeffresc/maxmind-geoip-authz

go 1.24.4

require (
        github.com/oschwald/geoip2-golang v1.11.0
        gopkg.in/yaml.v2 v2.4.0
        github.com/spf13/viper v0.0.0
)

replace github.com/spf13/viper => ./viper

require (
	github.com/oschwald/maxminddb-golang v1.13.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
)
