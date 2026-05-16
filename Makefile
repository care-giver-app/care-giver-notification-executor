default: invoke

EVENT=event.json

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bootstrap main.go

zip:
	zip -j function.zip ./bootstrap

sam-build: build zip
	sam build

invoke: sam-build
	sam local invoke \
	--invoke-image amazon/aws-lambda-provided:al2 \
	--event events/${EVENT} \
	--config-env local
	
start-api: sam-build
	sam local start-api --docker-network care-giver-infra_default

deploy-dev: sam-build
	sam deploy --config-env dev

local-atdd: 
	cd atdd && go test

test:
	go test -short -coverprofile cover.out ./...

test-report: test
	go tool cover -html=cover.out

lint: 
	golangci-lint run