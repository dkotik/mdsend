-include .env
export

default:
	@clear
	@echo ":::::::::::::::::::::::::::::::::::::::::"
	@go test ./...
	@#cd internal/mime && go test ./...
	@#cd sender && go test ./...
	@#cd queue && go test ./...
	@#cd cmd/mdsend && go test ./...
	@echo "::::::::::::::::::::::::::::::::::::"
live:
	@clear
	# @rm sender/smtp/testdata/live_test.lock
	@cd sender/smtp && go test ./... -v
	# @rm sender/mailgun/testdata/live_test.lock
	@cd sender/mailgun && go test ./... -v
build:
	goreleaser release --snapshot --rm-dist
install:
	cd ./cmd/gui && go build -trimpath -o ~/.local/bin/mdsend
	chmod +x ~/.local/bin/mdsend
