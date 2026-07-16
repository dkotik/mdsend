-include .env
export

default:
	@clear
	@echo ":::::::::::::::::::::::::::::::::::::::::"
	@go test ./...
	@#cd internal/template && go test ./...
	@#cd queue && go test ./...
	@#cd cmd/mdsend && go test ./...
	@echo "::::::::::::::::::::::::::::::::::::"
live:
	@clear
	# @rm mailer/smtp/testdata/live_test.lock
	# @rm mailer/mailgun/testdata/live_test.lock
	@cd mailer && go test ./... -v -count=1
build:
	goreleaser release --snapshot --rm-dist
update:
	@echo Updating project test data golden files...
	@cd internal/template && go test . -update
install:
	cd ./cmd/mdsend && go build -trimpath -o ~/.local/bin/mdsend
	chmod +x ~/.local/bin/mdsend
