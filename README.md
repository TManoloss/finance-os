# Personal Finance OS

Sistema de acompanhamento e decisão financeira pessoal com Open Finance e explicações opcionais por IA. Não movimenta dinheiro nem executa investimentos.

## Como iniciar o ambiente local

Para rodar o banco de dados e as ferramentas de desenvolvimento, siga os passos abaixo:

1.  **Configurar variáveis de ambiente:**
    Copie o arquivo de exemplo e ajuste se necessário:
    ```bash
    cp .env.example .env
    ```

2.  **Iniciar o Docker Compose:**
    ```bash
    docker compose up -d
    ```

O banco de dados PostgreSQL estará disponível na porta `5432` (ou a definida em `.env`) e o Adminer para visualização na porta `8080`.

## Estrutura do Projeto

- `backend/`: API em Go com Echo.
- `agents/`: Serviço de agentes em Python com FastAPI.
- `web/`: Dashboard em Next.js.
- `mobile/`: App em Flutter.

## Documentação canônica

- [`CONTEXT.md`](CONTEXT.md): vocabulário do domínio.
- [`docs/FINANCEOS-VALIDATION-PLAN.md`](docs/FINANCEOS-VALIDATION-PLAN.md): requisitos, critérios e evidências.
- [`docs/FINANCEOS-HANDOFF.md`](docs/FINANCEOS-HANDOFF.md): arquitetura, fluxos, deploy, testes e estado para continuar em outra sessão.
