
# Build binaries for all OS/Architectures
snapshot:
	@goreleaser \
		--rm-dist \
		--snapshot

release:
	@go get ./util/github-release-notes/
	@goreleaser \
		--rm-dist \
		--release-notes <(github-release-notes)


start-mongodb:
	@docker rm -f ktl-mongodb-test > /dev/null 2>&1 || echo
	@docker run -d --name ktl-mongodb-test -p 27017:27017 docker.io/mongo:3.5.13 > /dev/null
