PHONY: docker_build

docker_build:
	@rm webook || true
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
	@docker rmi -f riiceball/webook:v0.0.1
	@docker build -t riiceball/webook:v0.0.1 .