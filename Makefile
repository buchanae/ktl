
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
