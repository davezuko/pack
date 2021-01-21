default:
	go build  "-ldflags=-s -w" ./cmd/pack

install:
	go install  "-ldflags=-s -w" ./cmd/pack

publish:
	@test master = "`git rev-parse --abbrev-ref HEAD`" || (echo "Refusing to publish from non-master branch `git rev-parse --abbrev-ref HEAD`" && false)

.PHONY: npm
npm:
	cd npm; rm -rf dist
	cd npm; npx tsc --outDir dist/cjs --module commonjs
	cd npm; npx tsc --outDir dist/ejs --module esnext