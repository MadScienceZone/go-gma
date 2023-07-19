DIRS=map-console map-update preset-update server upload-presets coredb

binaries:
	@echo "(run 'make help' for instructions)"
	@for d in $(DIRS); do \
		echo "building $$d"; \
		(cd cmd/$$d && go build); \
	done

all: binaries manpages

clean:
	@for d in $(DIRS); do \
		echo "removing $$d binary"; \
		(cd cmd/$$d && rm -f $$d); \
	done

manpages:
	(cd man && $(MAKE))

test:
	go test ./...

telemetry:
	@echo "Building server with telemetry instrumentation enabled..."
	@echo "WARNING: This feature is not fully implemented yet and is not ready for production use."
	(cd cmd/server && go build -tags instrumentation)

help:
	@echo "To remove compiled binaries:"
	@echo "   $$ make clean"
	@echo ""
	@echo "To run unit tests:"
	@echo "   $$ make test"
	@echo ""
	@echo "To compile binaries: (no telemetry)"
	@echo "   $$ make binaries (the default)"
	@echo ""
	@echo "To compile binaries with telemetry agent integrated:"
	@echo "   $$ make telemetry"
	@echo ""
	@echo "To format manpages to PDF:"
	@echo "   $$ make manpages"
