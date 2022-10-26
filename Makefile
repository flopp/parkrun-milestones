.PHONY: build
build:
	go build -o parkrun-events cmd/events/main.go
	go build -o parkrun-milestones cmd/milestones/main.go

.PHONY: vet
vet:
	go vet ./...
