.PHONY: build
build: 
	mkdir -p .bin
	go build -o .bin/parkrun-events cmd/events/main.go
	go build -o .bin/parkrun-milestones cmd/milestones/main.go
	go build -o .bin/parkrun-runstats cmd/runstats/main.go
	go build -o .bin/parkrun-webgen cmd/webgen/main.go
	go build -o .bin/parkrun-year cmd/year/main.go
	go build -o .bin/parkrun-people cmd/people/main.go

.PHONY: vet
vet:
	go vet ./...

.PHONY: run-webgen
run-webgen:
	go run cmd/webgen/main.go -outdir html -country germany

.PHONY: run
run:
	@go run cmd/runstats/main.go -fancy dietenbach
