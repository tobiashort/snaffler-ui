BIN=snaffler-ui.exe

default: build

help:
	@echo "Targets:"
	@echo "  build (default)"
	@echo "  clean"
	@echo "  help"

build:
	GOOS=windows GOARCH=amd64 go build -o $(BIN)

clean:
	rm -f $(BIN)
