# Observabilidade com OpenTelemetry

Este projeto demonstra a implementação de observabilidade usando OpenTelemetry em dois servidores Go.

## Arquitetura

O projeto inclui:

- **Server A**: Servidor na porta 8081
- **Server B**: Servidor de clima na porta 8080
- **OpenTelemetry Collector**: Coleta e processa telemetria
- **Jaeger**: Visualização de traces
- **Prometheus**: Coleta de métricas
- **Grafana**: Dashboard para visualização

## Como executar

### 1. Pré-requisitos

Certifique-se de ter instalado:

- Docker
- Docker Compose

### 2. Configurar variáveis de ambiente

**⚠️ IMPORTANTE**: Este passo é essencial para o funcionamento correto do sistema!

Copie o arquivo de exemplo e configure sua chave da API:

```bash
cp env.example .env
```

Edite o arquivo `.env` e adicione sua chave da Weather API:

```
WEATHER_API_KEY=sua_chave_aqui
```

**Nota**: Para obter uma chave gratuita, acesse [WeatherAPI.com](https://www.weatherapi.com/)

**Problema comum**: Se você receber dados mockados ("Test City", temperatura 25°C), verifique se:

1. O arquivo `.env` foi criado corretamente
2. A `WEATHER_API_KEY` está configurada
3. Os containers foram rebuildados após a configuração

### 3. Subir o ambiente Docker

```bash
# Construir e iniciar todos os serviços
docker-compose up -d

# Verificar se todos os containers estão rodando
docker ps
```

### 4. Verificar status dos serviços

```bash
# Ver logs de todos os serviços
docker-compose logs

# Ver logs de um serviço específico
docker-compose logs server-a
docker-compose logs server-b
docker-compose logs otel-collector
```

### 5. Acessar as aplicações

- **Server A**: http://localhost:8081
- **Server B**: http://localhost:8080
- **Jaeger UI**: http://localhost:16686
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

### 6. Testar o fluxo completo

#### Teste 1: Fluxo completo (Server-A → Server-B)

```bash
curl -X POST http://localhost:8081/input \
  -H "Content-Type: application/json" \
  -d '{"cep":"01001000"}'
```

**Resposta esperada:**

```json
{ "city": "São Paulo", "temp_C": 22.5, "temp_F": 72.5, "temp_K": 295.65 }
```

**Nota**: Os valores de temperatura são reais obtidos da WeatherAPI e variam conforme a localização e horário.

#### Teste 2: Server B diretamente

```bash
curl -X POST http://localhost:8080/weather \
  -H "Content-Type: application/json" \
  -d '{"cep":"01001000"}'
```

#### Teste 3: CEP inválido

```bash
curl -X POST http://localhost:8081/input \
  -H "Content-Type: application/json" \
  -d '{"cep":"123"}'
```

### 7. Explorar a observabilidade

#### Jaeger (Traces)

1. Acesse http://localhost:16686
2. Selecione o serviço "server-a" ou "server-b"
3. Clique em "Find Traces" para ver os traces das requisições

#### Prometheus (Métricas)

1. Acesse http://localhost:9090
2. Vá em "Status" → "Targets" para ver os endpoints monitorados
3. Use a aba "Graph" para executar queries PromQL

#### Grafana (Dashboards)

1. Acesse http://localhost:3000
2. Login: admin/admin
3. Configure as datasources:
   - Prometheus: http://prometheus:9090
   - Jaeger: http://jaeger:16686
4. Crie dashboards para visualizar métricas e traces

## Observabilidade

### Traces

- Visualize traces distribuídos no Jaeger UI
- Cada requisição HTTP gera spans automáticos
- Spans customizados para operações específicas

### Métricas

- Métricas HTTP automáticas (duração, tamanho, status)
- Métricas customizadas do OpenTelemetry
- Visualização no Prometheus e Grafana

### Logs

- Logs estruturados via OpenTelemetry
- Correlação com traces via trace ID

## Estrutura do Projeto

```
.
├── server_A/                 # Servidor A
│   ├── main.go
│   ├── inputHandler.go
│   ├── otel.go
│   └── Dockerfile
├── server_B/                 # Servidor B
│   ├── main.go
│   ├── otel.go
│   └── Dockerfile
├── docker-compose.yaml       # Orquestração dos containers
├── otel-collector-config.yaml # Configuração do coletor
├── prometheus.yml           # Configuração do Prometheus
├── go.mod                   # Dependências Go
└── README.md
```

## Troubleshooting

### Problemas comuns

#### 1. Porta já em uso

```bash
# Verificar quais portas estão em uso
lsof -i :8080
lsof -i :8081
lsof -i :9090
lsof -i :3000
lsof -i :16686

# Parar containers conflitantes
docker stop $(docker ps -q)
```

#### 2. Containers não iniciam

```bash
# Ver logs detalhados
docker-compose logs

# Reconstruir containers
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

#### 3. Problemas de rede Docker

```bash
# Verificar redes Docker
docker network ls

# Limpar redes não utilizadas
docker network prune
```

#### 4. Problemas com OpenTelemetry Collector

```bash
# Verificar configuração
docker-compose logs otel-collector

# Reiniciar apenas o collector
docker-compose restart otel-collector
```

#### 5. Sistema retornando dados mockados

Se o sistema retornar `{"city":"Test City","temp_C":25,"temp_F":77,"temp_K":298.15}`, significa que a API key não está configurada:

```bash
# Verificar se o arquivo .env existe e tem a chave
cat .env

# Se não existir, criar o arquivo
cp env.example .env

# Rebuildar os containers para carregar a nova configuração
docker-compose down
docker-compose build --no-cache server-b
docker-compose up -d
```

### Comandos úteis

```bash
# Parar todos os serviços
docker-compose down

# Parar e remover volumes
docker-compose down -v

# Ver status dos containers
docker-compose ps

# Executar comando em container específico
docker-compose exec server-a sh
docker-compose exec server-b sh
```

## Desenvolvimento Local

Para executar sem Docker:

```bash
# Terminal 1 - Server A
cd server_A
go run .

# Terminal 2 - Server B
cd server_B
go run .
```

## O que foi implementado

### ✅ Funcionalidades

- **Server-A**: Servidor que recebe requisições e encaminha para o Server-B
- **Server-B**: Serviço de weather que processa CEPs e retorna informações de temperatura **reais**
- **APIs Externas**: Integração com ViaCEP (localização) e WeatherAPI (temperatura)
- **OpenTelemetry**: Instrumentação completa com traces, métricas e logs
- **Jaeger**: Visualização de traces distribuídos
- **Prometheus**: Coleta e armazenamento de métricas
- **Grafana**: Dashboards para visualização de dados
- **Docker Compose**: Orquestração completa do ambiente

### 🔧 Configurações

- **Traces**: Configurados automaticamente via OpenTelemetry
- **Métricas**: HTTP metrics automáticas + métricas customizadas
- **Logs**: Logs estruturados com correlação de traces
- **Networking**: Comunicação entre serviços via Docker network
- **Volumes**: Persistência de dados para Grafana

## APIs Externas Utilizadas

### ViaCEP

- **URL**: https://viacep.com.br/ws/{cep}/json/
- **Função**: Obter localidade (cidade) a partir do CEP
- **Exemplo**: CEP "13064722" → "Campinas"

### WeatherAPI

- **URL**: http://api.weatherapi.com/v1/current.json
- **Função**: Obter temperatura atual da cidade
- **Requisito**: API Key gratuita em https://www.weatherapi.com/
- **Exemplo**: "Campinas" → 17.1°C

## Tecnologias Utilizadas

- **Go**: Linguagem de programação
- **OpenTelemetry**: Padrão de observabilidade
- **Docker**: Containerização
- **Jaeger**: Visualização de traces
- **Prometheus**: Coleta de métricas
- **Grafana**: Dashboards
- **HTTP**: Comunicação entre serviços
