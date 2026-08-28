# Roteamento de Subagentes

Estas regras se aplicam a criação de funcionalidades, testes em dispositivo e depuração neste projeto, desde que as ferramentas e modelos estejam disponíveis e não conflitem com instruções de maior prioridade.

## Contexto obrigatório

- Antes de alterar domínio ou terminologia, leia `CONTEXT.md`; ele é o glossário canônico e não contém decisões técnicas.
- Antes de implementar ou validar requisitos FinanceOS, leia `docs/FINANCEOS-VALIDATION-PLAN.md`; ele é a fonte oficial de requisitos, critérios e evidências.
- Antes de continuar trabalho existente, publicar no Render, mexer no app Flutter ou testar no Samsung, leia `docs/FINANCEOS-HANDOFF.md`; ele registra arquitetura, fluxos, comandos, estado Git, limitações e próximo trabalho.

## Papéis

- O agente principal atua como Engenheiro e Arquiteto Principal em `gpt-5.6-sol` com esforço médio: define o plano estrutural, limites de arquitetura, contratos e decisões complexas de lógica.
- Trabalho repetitivo — boilerplate, edições mecânicas e scripts de testes unitários/UI — deve ser delegado a um subagente `gpt-5.6-luna` com esforço máximo.
- O agente principal revisa toda entrega delegada, integra somente mudanças dentro do escopo e executa a verificação final.

## Fluxo

1. Antes de implementar, o agente principal identifica contratos, riscos, arquivos afetados e critérios verificáveis de conclusão.
2. Se houver trabalho repetitivo independente, abre um Luna Max com escopo fechado, arquivos permitidos e comando de validação. A delegação termina somente quando o subagente entrega o código e o resultado do teste solicitado.
3. Em testes CLI no dispositivo, o agente principal preserva o log original localmente, remove credenciais, tokens e dados pessoais na fronteira de coleta e envia o restante a um Luna Max com a instrução: “Limpe este log de depuração, remova os avisos do sistema operacional e retorne apenas a stack trace exata do erro fatal”.
4. O agente principal confronta a stack trace limpa com o log original sanitizado antes de decidir a correção.
5. Uma tentativa automática conta somente quando o Luna Max altera o código e executa o teste que reproduz a falha. Após duas tentativas malsucedidas para a mesma causa, o agente principal assume a depuração lógica profunda.
6. Se subagentes ou o modelo solicitado estiverem indisponíveis, o agente principal informa a limitação e executa o menor fluxo seguro que mantenha o trabalho avançando.

## Limites

- Delegação não transfere decisões de arquitetura, autorização para ações destrutivas nem responsabilidade pela validação final.
- Subagentes recebem apenas o contexto e os arquivos necessários ao subproblema.
- Alterações locais do Usuário e a separação entre o repositório raiz e `mobile/` devem ser preservadas.
