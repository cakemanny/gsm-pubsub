
gsm-pubsub: go.mod go.sum main.go
	go build .

.PHONY: test
test:
	go test .

.PHONY: docker
docker:
	docker build -t cakemanny/gsm-pubsub:latest .

