APP_NAME = "docs-server"

docker:
	docker build -t ${APP_NAME} . --platform linux/amd64