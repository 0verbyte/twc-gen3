.PHONY: build
build: web server

.PHONY: web
web:
	cd web && npm install && npm run build

.PHONY: server
server:
	go build
