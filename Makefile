DIRS=map-console map-update preset-update server upload-presets coredb session-stats image-audit roll markup
DESTDIR=/opt/gma

binaries:
	@echo "(run 'make help' for instructions)"
	@for d in $(DIRS); do \
		echo "building $$d"; \
		(cd cmd/$$d && go build); \
	done

all: binaries manpages

# for internal testing purposes only
refresh-alpha:
	install -d $(DESTDIR)/bin
	@echo "updating alpha server binary to $(DESTDIR)/bin..."
	mv -v $(DESTDIR)/bin/server-alpha $(DESTDIR)/bin/server-alpha.$$
	install cmd/server/server $(DESTDIR)/bin/server-alpha

refresh:
	install -d $(DESTDIR)/bin
	@echo "Installing binaries to $(DESTDIR)/bin..."
	@for d in $(DIRS); do \
		install cmd/$$d/$$d $(DESTDIR)/bin; \
	done
	@echo "Installing manpages to $(DESTDIR)/man..."
	(cd man && DESTDIR="$(DESTDIR)" $(MAKE) install)

install: refresh
	install -d $(DESTDIR)/var
	@echo "Installing sample server files in $(DESTDIR)/var..."
	install -m 0600 cmd/server/sample.* $(DESTDIR)/var
	@echo "NOTE: add $(DESTDIR)/man to your MANPATH variable"
	@echo "NOTE: add $(DESTDIR)/bin to your PATH variable"
	@echo "NOTE: customize your server files in $(DESTDIR)/var"

clean:
	@for d in $(DIRS); do \
		echo "removing $$d binary"; \
		(cd cmd/$$d && rm -f $$d); \
	done

manpages:
	(cd man && $(MAKE))

test:
	go test ./... && go vet ./...

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
	@echo ""
	@echo "To install into $(DESTDIR):"
	@echo "   $$ make install"
