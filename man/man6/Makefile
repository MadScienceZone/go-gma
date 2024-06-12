all: map-console.6.pdf map-update.6.pdf preset-update.6.pdf server.6.pdf upload-presets.6.pdf coredb.6.pdf session-stats.6.pdf image-audit.6.pdf roll.6.pdf markup.6.pdf

install:
	@echo "Installing manpages to $(DESTDIR)/man/man6..."
	install -d $(DESTDIR)/man/man6
	install -m 644 *.6 $(DESTDIR)/man/man6

markup.6.pdf: markup.6
	gma fmtman < $< | groff -man | ps2pdf - $@

roll.6.pdf: roll.6
	gma fmtman < $< | groff -man | ps2pdf - $@

image-audit.6.pdf: image-audit.6
	gma fmtman < $< | groff -man | ps2pdf - $@

session-stats.6.pdf: session-stats.6
	gma fmtman < $< | groff -man | ps2pdf - $@

coredb.6.pdf: coredb.6
	gma fmtman < $< | groff -man | ps2pdf - $@

map-console.6.pdf: map-console.6
	gma fmtman < $< | groff -man | ps2pdf - $@

map-update.6.pdf: map-update.6
	gma fmtman < $< | groff -man | ps2pdf - $@

preset-update.6.pdf: preset-update.6
	gma fmtman < $< | groff -man | ps2pdf - $@

server.6.pdf: server.6
	gma fmtman < $< | groff -man | ps2pdf - $@

upload-presets.6.pdf: upload-presets.6
	gma fmtman < $< | groff -man | ps2pdf - $@
