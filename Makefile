-include .env
export

default:
	@clear
	@echo ":::::::::::::::::::::::::::::::::::::::::"
	@cd internal/mime && go test ./...
	@cd sender && go test ./...
	@cd queue && go test ./...
	@echo "::::::::::::::::::::::::::::::::::::"
	@# go test -v ./locks/ -args --integration
	@# go test -v ./userinterface/bubbletea/...
	@# go generate ./...
	@# go test ./tests/ -run TestImport
	@# go test ./tests/
	@# go run ./cmd/gui/ tests/data/testemail.md
live:
	@clear
	@rm sender/smtp/testdata/live_test.lock
	@cd sender/smtp && go test ./... -v
	# @rm sender/mailgun/testdata/live_test.lock
	# @cd sender/mailgun && go test ./... -v
build:
	goreleaser release --snapshot --rm-dist
install:
	cd ./cmd/gui && go build -trimpath -o ~/.local/bin/mdsend
	chmod +x ~/.local/bin/mdsend
development:
	go get github.com/gotk3/gotk3/gtk
	sudo apt-get install appmenu-gtk3-module libgtk-3-dev libcairo2-dev libglib2.0-dev
cleanup:
	rm ~/.cache/mdsend.lock/*.lock
