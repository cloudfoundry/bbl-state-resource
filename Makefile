# simple makefile to build bbl-state-resource docker container

.DEFAULT_GOAL := help

GO_SRCS = $(shell find . -type f -name '*.go')

test: ## run all the tests
	ginkgo -r -race --randomizeAllSpecs

docker: $(GO_SRCS) ## rebuild your local docker container, test all the builds
	docker build --rm -t bbl-state-resource .
	touch .docker_built # sentinel file indicating time we last built

.docker_built: docker
push: .docker_built ## commit and send it to it's canonical location within dockerhub
	# git push
	docker tag bbl-state-resource cfinfrastructure/bbl-state-resource
	docker push cfinfrastructure/bbl-state-resource

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

