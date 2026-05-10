BINARY := soundwires
WEB_SRC := $(shell find web/src -type f 2>/dev/null) web/index.html web/package.json web/vite.config.ts

.PHONY: build dev-go dev-web clean install

build: web/dist
	go build -o $(BINARY) .
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 \
		go build -buildvcs=false -o $(BINARY)-armhf .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		go build -buildvcs=false -o $(BINARY)-arm64 .


web/dist: web/node_modules $(WEB_SRC)
	cd web && npm run build

web/node_modules: web/package.json
	cd web && npm install
	@touch web/node_modules

install: web/dist
	go install .

dev-go:
	go run . --port 8080

dev-web:
	cd web && npm run dev

clean:
	rm -rf web/dist $(BINARY)
