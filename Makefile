.PHONY: build
build: 
	go build -o parkrun-events cmd/events/main.go
	go build -o parkrun-milestones cmd/milestones/main.go
	go build -o parkrun-runstats cmd/runstats/main.go
	go build -o parkrun-webgen cmd/webgen/main.go
	go build -o parkrun-year cmd/year/main.go

.PHONY: vet
vet:
	go vet ./...

.PHONY: run-webgen
run-webgen:
	go run cmd/webgen/main.go -outdir html -country germany