all: gma-go-map-console.6.pdf gma-go-map-update.6.pdf gma-go-preset-update.6.pdf gma-go-server.6.pdf gma-go-upload-presets.6.pdf gma-go-coredb.6.pdf gma-go-session-stats.6.pdf gma-go-image-audit.6.pdf gma-go-roll.6.pdf gma-go-markup.6.pdf

install:
	@echo "Installing manpages to $(DESTDIR)/man/man6..."
	install -d $(DESTDIR)/man/man6
	install -m 644 *.6 $(DESTDIR)/man/man6

gma-go-markup.6.pdf: gma-go-markup.6
	gma fmtman < $< | groff -man | ps2pdf - $@

gma-go-roll.6.pdf: gma-go-roll.6
	gma fmtman < $< | groff -man | ps2pdf - $@

gma-go-image-audit.6.pdf: gma-go-image-audit.6
	gma fmtman < $< | groff -man | ps2pdf - $@

gma-go-session-stats.6.pdf: gma-go-session-stats.6
	gma fmtman < $< | groff -man | ps2pdf - $@

gma-go-coredb.6.pdf: gma-go-coredb.6
	gma fmtman < $< | groff -man | ps2pdf - $@

gma-go-map-console.6.pdf: gma-go-map-console.6
	gma fmtman < $< | groff -man | ps2pdf - $@

gma-go-map-update.6.pdf: gma-go-map-update.6
	gma fmtman < $< | groff -man | ps2pdf - $@

gma-go-preset-update.6.pdf: gma-go-preset-update.6
	gma fmtman < $< | groff -man | ps2pdf - $@

gma-go-server.6.pdf: gma-go-server.6
	gma fmtman < $< | groff -man | ps2pdf - $@

gma-go-upload-presets.6.pdf: gma-go-upload-presets.6
	gma fmtman < $< | groff -man | ps2pdf - $@
