.PHONY: build
build: 
	go build -o parkrun-events cmd/events/main.go
	go build -o parkrun-milestones cmd/milestones/main.go
	go build -o parkrun-runstats cmd/runstats/main.go
	go build -o parkrun-webgen cmd/webgen/main.go
	go build -o parkrun-year cmd/year/main.go
	go build -o parkrun-graphs cmd/graphs/main.go
	go build -o parkrun-people cmd/people/main.go
	go build -o freiburg-temp cmd/freiburgtemp/main.go

.PHONY: vet
vet:
	go vet ./...

.PHONY: run-webgen
run-webgen:
	go run cmd/webgen/main.go -outdir html -country germany

.PHONY: run
run:
	@go run cmd/runstats/main.go -fancy dietenbach
	@echo
	@go run cmd/people/main.go dietenbach people.html
	@scp people.html echeclus.uberspace.de:/var/www/virtual/floppnet/freiburg.run/
	@scp people.html echeclus.uberspace.de:/var/www/virtual/floppnet/fraig.de/
