all: connclose

connclose:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false -o connclose .
.PHONY: connclose

image: connclose
	docker build -t connclose:latest .

image-load:
	kind load docker-image connclose:latest

deploy:
	kubectl apply -f secret.yml -f deployment.yml -f network.yml -f network.allow.yml

start:
	bash start.sh

down:
	kind delete cluster

restart:
	kubectl rollout restart deployment bot

update: image image-load restart

allow:
	kubectl apply -f network.allow.yml

deny:
	kubectl apply -f network.deny.yml

wait:
	kubectl rollout status deployment bot

logs:
	kubectl logs -l app=bot

follow:
	kubectl logs -f -l app=bot

up: start image image-load deploy wait
