.PHONY: cli-run cli-build cli-test www-dev www-build www-lint

CLI_ARGS ?=

cli-run:
	cd cli && go run ./cmd/warhol $(CLI_ARGS)

cli-build:
	cd cli && go build ./...

cli-test:
	cd cli && go test ./...

www-dev:
	cd www && pnpm dev

www-build:
	cd www && pnpm build

www-lint:
	cd www && pnpm lint

