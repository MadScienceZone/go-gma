DIRS=map-console map-update preset-update server upload-presets

all: binaries manpages

binaries:
	for d in $(DIRS); do \
		echo "building $$d"; \
		(cd cmd/$$d && go build); \
	done

manpages:
	(cd man && $(MAKE))

test:
	go test ./...

telemetry:
	@echo "Building server with telemetry instrumentation enabled..."
	@echo "WARNING: This feature is not fully implemented yet and is not ready for production use."
	(cd cmd/server && go build -tags instrumentation)