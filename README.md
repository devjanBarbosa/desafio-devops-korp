# Desafio DevOps - Projeto Korp

Solução completa de engenharia de infraestrutura, observabilidade e automação para o desafio técnico da Korp. O projeto implementa uma arquitetura de microsserviços orientada a microrredes isoladas, telemetria declarativa (Golden Signals) e provisionamento idempotente via Ansible encapsulado.

---

## 🏛️ Visão Geral da Arquitetura

A solução é composta por 4 componentes conteinerizados sob uma malha de rede bridge dedicada (`korp-network`), operando sob o princípio do menor privilégio e isolamento de borda:

              +-------------------------------------------------------+
              |                     HOST MACHINE                      |
              +-------------------------------------------------------+
                       | :80             | :9090           | :3000
                       v                 v                 v
             +-------------------+ +---------------+ +---------------+
             |    nginx-proxy    | |  prometheus   | |    grafana    |
             |  (Reverse Proxy)  | |  (v2.50.1)    | |   (10.4.1)    |
             +-------------------+ +---------------+ +---------------+
                       |                        |                 |
  [Internal Bridge]    | http-server:8080       | Scrape :8080    | Read Datasource
  (korp-network)       v                        v                 v
             +-------------------------------------------------------+
             |               http-server-projeto-korp                |
             |          (Go 1.22 Native API + Prometheus SDK)        |
             +-------------------------------------------------------+

### Componentes e Decisões de Design

1. **API Go (`http-server-projeto-korp`):**
   * **Multi-stage Build:** Compilação em `golang:1.22-alpine` gerando um binário estático puro (`CGO_ENABLED=0`, `-ldflags="-w -s"`), empacotado sobre imagem enxuta `alpine:3.19` (~15MB total).
   * **Segurança:** Execução sob usuário sem privilégios (`appuser`, UID 1000).
   * **Telemetria Nativa:** Integração com o SDK oficial `client_golang` expondo as métricas exigidas:
     * `http_server_up` (Gauge): Indicador de vivacidade do serviço.
     * `http_requests_total` (Counter com labels `path` e `status`): Volume cumulativo de requisições.
     * Métricas de runtime do Go (Heap alocada, Goroutines ativas, uso de CPU e GC).
   * **Timeouts Defensivos:** `ReadHeaderTimeout`, `ReadTimeout` e `WriteTimeout` explicitamente configurados contra ataques de exaustão de conexões (Slowloris).

2. **Reverse Proxy (`nginx`):**
   * Ponto de terminação e exposição pública na porta `80`.
   * Bloqueio de métodos HTTP não autorizados (somente `GET` permitido em `/projeto-korp`).
   * Headers de encaminhamento transparentes (`X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto`).
   * Desativação de buffers para métricas em tempo real (`proxy_buffering off`).

3. **Monitoramento & Métricas (`prometheus`):**
   * Coletor com *scrape interval* configurado em 5s direcionado ao endpoint interno `/metrics`.
   * Retenção de TSDB ajustada para 7 dias (`--storage.tsdb.retention.time=7d`).

4. **Visualização Declarativa (`grafana`):**
   * **Provisioning as Code (IaC):** Datasource Prometheus e Dashboard SRE provisionados automaticamente na inicialização via arquivos YAML/JSON, sem necessidade de configuração manual via interface.
   * Painéis customizados cobrindo:
     * **Golden Signals:** Service Status (Gauge), Total de Requisições (Counter), Throughput Atual em req/s (`rate()`) e Taxa de Erro HTTP 4xx/5xx.
     * **Tráfego HTTP:** Séries temporais de RPS segregadas por rota e código de status, além de gráfico de distribuição.
     * **Runtime do Go:** Monitoramento de Goroutines ativas e alocação de memória Heap (`HeapAlloc` vs `HeapSys`).

5. **Infraestrutura Imutável:**
   * Todas as configurações (`nginx.conf`, `prometheus.yml`, dashboards) são incorporadas (*baked*) diretamente em Dockerfiles dedicados de cada serviço, eliminando fragilidades de *bind mounts* e inconsistências de paths entre SOs.

---

## 🛡️ Resiliência e Operabilidade em Produção

* **Healthchecks Nativos & Orquestração Determinística:** A API Go implementa healthcheck interno via `wget`. Os serviços dependentes (`nginx` e `prometheus`) utilizam `condition: service_healthy` no Docker Compose, prevenindo falhas de conexão (*502 Bad Gateway*) durante o boot da stack.
* **Política de Retenção de Logs:** Driver de log `json-file` com limites estritos (`max-size: 10m`, `max-file: 3`) em todos os containers para impedir a saturação do disco por logs descontrolados.
* **Simulação de Carga Automatizada (`scripts/load-generator.py`):** Script idempotente que injeta requisições de sucesso e cenários de erro (HTTP 404) para validação ativa dos alertas e gráficos do Grafana.

---

## 🤖 Automação com Ansible (Control Node Containerizado)

Para garantir reprodutibilidade estrita (Zero Drift) sem exigir dependências locais no host do avaliador, a automação com Ansible opera via **Runner Containerizado** (`ansible/Dockerfile`):

* **Imagem Base:** Alpine Linux 3.20 com `ansible-core 2.16`, biblioteca `docker 7.1.0` e a coleção oficial `community.docker 3.10.0` travadas.
* **Socket Binding:** Comunicação direta com o daemon do Docker através da montagem de `/var/run/docker.sock`.
* **Playbook Idempotente (`ansible/playbooks/deploy.yml`):**
  1. Validação de conectividade com o Docker Daemon.
  2. Asserção estrita de integridade de todos os arquivos de configuração necessários.
  3. Build e subida da stack completa via `community.docker.docker_compose_v2`.
  4. Teste de fumaça (*healthcheck*) com política de retry contra a borda do Nginx.
  5. Disparo do gerador de telemetria e carga sintética.

---

## 🔄 Esteira de Integração Contínua (CI/CD Pipeline)

O repositório possui integração contínua via **GitHub Actions** (`.github/workflows/ci.yml`), disparada a cada push ou pull request na branch `main`:

* Setup do ambiente com Docker Buildx.
* Compilação das imagens imutáveis e inicialização da stack completa via Docker Compose.
* Validação de integridade dos containers e execução do teste de carga simulado.
* Coleta de logs e status da infraestrutura para inspeção contínua.

---

## 🚀 Como Executar o Projeto

### Pré-requisitos
* Docker e Docker Compose instalados e em execução.

### Método 1: Provisionamento Automatizado via Ansible (Recomendado)
Execute o script facilitador na raiz do repositório:

powershell

# No Windows (PowerShell):
.\run-ansible.ps1

# Ou execute manualmente o container do Ansible:
docker build -t korp-ansible-runner ./ansible
docker run --rm -it -v //var/run/docker.sock:/var/run/docker.sock -v "${PWD}:/workspace" korp-ansible-runner playbooks/deploy.yml

### Método 2: Execução Direta via Docker Compose

# Bash
docker compose up -d --build

### 🧪 Verificação Manual e Teste de Carga

Para testar o endpoint manualmente e acompanhar a variação em tempo real no Grafana:

# Requisição de Sucesso (HTTP 200)
curl.exe -i http://localhost/projeto-korp

# Disparar teste de carga sintético
python scripts/load-generator.py
