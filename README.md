# Client Server Api em GO

Esta é uma atividade do curso de pós graduação Go Expert da Full Cicle.

## Descrição

Trata-se um servidor que disponibiliza uma API para consulta da cotação do dólar americano.
Para isso acessa uma outra API externa que disponibiliza a cotação do dólar. Após a consulta, o servidor retorna a cotação e salva o registro recebido da API externa no banco de dados (SQLite).

No lado cliente, é um programa que acessa a API do servidor e salva a cotação do dólar em um arquivo `cotacao.txt`.

Temos ainda o atendimento de alguns requisitos como utilização de contextos e logs de erros.

## Como rodar

### Servidor

```bash
go run ./server/server.go
```

### Cliente

```bash
go run ./client/client.go
```

## Endpoints

### GET /cotacao

Retorna a cotação do dólar americano.

```bash
curl http://localhost:8080/cotacao
```

