default:
	go build  "-ldflags=-s -w" ./cmd/pack

publish:
	@test master = "`git rev-parse --abbrev-ref HEAD`" || (echo "Refusing to publish from non-master branch `git rev-parse --abbrev-ref HEAD`" && false)