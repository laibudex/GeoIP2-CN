GOOS=linux GOARCH=amd64 go build -o dist/ipip3mmdb
go build -o dist/ipip2mmdb main.go ip2cidr.go
go build -o dist/verify_ip verify/verify_ip.go
