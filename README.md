# Personal Finance OS

Sistema de gestão financeira pessoal e familiar com Open Finance e IA.

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
