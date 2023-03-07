DIRS=map-console map-update preset-update server upload-presets

all:
	for d in $(DIRS); do \
		echo "building $$d"; \
		(cd cmd/$$d && go build); \
	done

test:
	go test ./...
