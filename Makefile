.DEFAULT_GOAL := help

.PHONY: help
help: ## Show the available commands
	@printf "\033[33mUsage:\033[0m\n  make [target] [arg=\"val\"...]\n\n\033[33mTargets:\033[0m\n"
	@grep -E '^[-a-zA-Z0-9_\.\/]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[32m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: init
init: ## Initialize the application
	@printf "\033[32mInitializing the application...\033[0m\n"
	@cp ./docker/ais-receiver/.env.dist ./docker/ais-receiver/.env
	@make build

.PHONY: build
build: ## Build the docker image
	@printf "\033[32mBuilding docker image...\033[0m\n"
	@docker compose build

.PHONY: run
run: ## Start the application
	@printf "\033[32mStarting the application...\033[0m\n"
	@docker compose up -d

.PHONY: upload
upload: ## Uploads the Docker image to Docker Hub
	@printf "\033[32mUploading the Docker image to Docker Hub...\033[0m\n"
	@printf	"\033[33mBuilding the Docker image...\033[0m "
	@docker build -t auxority/ais-receiver .
	@printf "\033[32mPushing the image to Docker Hub...\033[0m\n"
	@docker push auxority/ais-receiver
	@printf "\033[32mDone!\033[0m\n"
