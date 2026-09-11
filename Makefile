.PHONY: help up down restart logs ps ansible test lint clean

help: ## Exibe os comandos disponíveis
@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

up: ## Sobe a stack com build imutável em background
docker compose up -d --build

down: ## Remove containers, redes e volumes da stack
docker compose down -v

restart: down up ## Reinicia completamente o ambiente

ps: ## Lista o status e healthcheck dos containers
docker compose ps

logs: ## Acompanha os logs de todos os serviços em tempo real
docker compose logs -f

ansible: ## Executa o provisionamento idempotente via runner containerizado
docker build -t korp-ansible-runner ./ansible
docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock -v "$$(pwd):/workspace" korp-ansible-runner playbooks/deploy.yml

test: ## Dispara o gerador de telemetria e carga sintética
python3 scripts/load-generator.py

lint: ## Executa validações estáticas de código Go e Docker Compose
cd app && go vet ./...
docker compose config --quiet

clean: down ## Limpeza completa de imagens residuais da Korp
docker rmi -f http-server-projeto-korp nginx-proxy prometheus grafana korp-ansible-runner 2>/dev/null || true