.PHONY: test
test:
	go test ./... -coverpkg=./... -coverprofile=coverage.out -count=1

.PHONY: cover 
cover:
	go tool cover -html=coverage.out
