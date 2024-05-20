PHONY: docker_build mock

docker_build:
	@rm webook || true
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
	@docker rmi -f riiceball/webook:v0.0.1
	@docker build -t riiceball/webook:v0.0.1 .

mock:
	@mockgen -package=redismocks -destination=./internal/repository/cache/redismocks/cmd.mock.go github.com/redis/go-redis/v9 Cmdable
	@go generate ./...
	@go mod tidy
