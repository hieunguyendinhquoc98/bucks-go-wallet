GO_VERSION = 1.20
GO_BUILD_FLAGS = -ldflags "-s -w"

winos:
	go env -w GOOS=windows

macos:
	go env -w GOOS=darwin

run: winos
	go run cmd/main.go

tidy:
	go mod tidy && go mod vendor

fmt:
	find . -iname '*.go' -not -path '*/vendor/*' -print0 | xargs -0 gofmt -s -w