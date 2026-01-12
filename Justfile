IMAGE := "kuberhealthy/dns-resolution-check"
TAG := "latest"

# Build the DNS resolution check container locally.
build:
	podman build -f Containerfile -t {{IMAGE}}:{{TAG}} .

# Run the unit tests for the DNS resolution check.
test:
	go test ./...

# Build the DNS resolution check binary locally.
binary:
	go build -o bin/dns-resolution-check ./cmd/dns-resolution-check
