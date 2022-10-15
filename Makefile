tripreporter: deps-server build-server deps-ui build-ui run

# deps

deps-server:
	go get -u "github.com/go-sql-driver/mysql"
	go get -u "github.com/joho/godotenv"
	go mod tidy

deps-ui:
	cd ui/; \
	npm i

# build

build-server:
	go build -o tripreporter .
	chmod +x tripreporter

build-ui: deps-ui
	cd ui/; \
	npm run build

# run production

run:
	./tripreporter

# run development (needs both dev and dev-ui separately)

dev-ui: deps-ui
	cd ui/; \
	npm run serve

dev-server: build-server
	./tripreporter -dev
