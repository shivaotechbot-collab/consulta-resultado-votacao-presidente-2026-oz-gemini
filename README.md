# Consulta de Resultado da Votação para Presidente 2026 — Osasco/SP

Aplicação desenvolvida com estética zen minimalista, arquitetura de alta performance e rigor técnico (TDD) para consulta e detalhamento dos resultados da eleição geral federal de 2026 em Osasco (SP).

## Arquitetura

1. **Frontend (Separado & Estático):**
   - HTML5 semântico, CSS3 com variáveis de design zen minimalista e JavaScript Vanilla assíncrono.
   - Servido diretamente pelo **Nginx** na porta `8080` (diretório `/var/www/consulta-votacao-oz`).
   - Apresentação de métricas gerais (eleitorado, presença, abstenções), cartões dos candidatos com barra percentual, formulário de pesquisa multifiltro (Zona, Seção, Candidato, Local) e modal de apuração detalhada por seção.

2. **Backend (Go / Golang):**
   - Desenvolvido em Go com cobertura total de testes unitários (`internal/data` e `internal/handler`).
   - Ao iniciar, gera um banco de dados **SQLite temporário em memória/disco efêmero**, cria índices de consulta rápida e importa o arquivo CSV de votos (`/home/hermes/projetos/referencia-dados-presidente/osasco-presidente.csv`).
   - API REST servida em `127.0.0.1:8089` e gerenciada como serviço pelo `systemd` (`votacao-oz.service`).

3. **Nginx (Proxy Reverso & Servidor Estático):**
   - Arquivos estáticos servidos em `/` com compressão Gzip e headers de segurança.
   - Rotas de API mapeadas em `/api/` com proxy reverso transparente para o binário Go.

## Estrutura do Projeto

```
consulta-resultado-votacao-presidente-2026-oz-gemini/
├── cmd/
│   └── server/
│       └── main.go           # Ponto de entrada, inicialização do SQLite temporário e servidor HTTP
├── internal/
│   ├── data/
│   │   ├── repository.go     # Acesso ao SQLite temporário e importação de CSV (Latin-1 -> UTF-8)
│   │   └── repository_test.go# Testes unitários do repositório (TDD)
│   ├── handler/
│   │   ├── handler.go        # Endpoints REST e CORS
│   │   └── handler_test.go   # Testes unitários dos handlers HTTP (TDD)
│   └── model/
│       └── model.go          # Estruturas de dados (Summary, CandidateTotal, SectionResult, etc.)
├── static/
│   ├── css/
│   │   └── style.css         # Design zen minimalista
│   ├── js/
│   │   └── app.js            # Lógica Vanilla JS de busca e renderização
│   └── index.html            # Estrutura HTML5 da aplicação
├── bin/
│   └── server                # Binário Go compilado
├── go.mod
├── go.sum
└── README.md
```

## Executando os Testes (TDD)

```bash
GOTOOLCHAIN=local go test -v ./...
```

## Gerenciamento do Serviço

- **Status do Backend:** `sudo systemctl status votacao-oz`
- **Reiniciar Backend:** `sudo systemctl restart votacao-oz`
- **Configuração Nginx:** `/etc/nginx/sites-available/consulta-votacao-oz`
- **Porta pública do sistema:** `http://<IP_DO_SERVIDOR>:8080` (ou `http://127.0.0.1:8080`)
