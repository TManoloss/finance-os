# Design Spec: Finance OS Mobile (Blueprint Edition)

**Data:** 10 de Maio de 2026  
**Status:** Em Revisão  
**Versão:** 1.0 (Mobile)

## 1. Visão Geral
Versão mobile do Personal Finance OS, trazendo a estética "Blueprint/Técnica" aprovada na Web para o ambiente Flutter. O app deve ser focado em telemetria financeira rápida, logs de transações e interação com a IA Pierre.

## 2. Design System (Mobile)

### 2.1 Paleta de Cores
*   **Background:** `#f4f1ea` (Bege Blueprint)
*   **Border:** `#000000` (Preto Sólido, 2px)
*   **Elevated/Surface:** `#e8e5de` (Cinza Técnico)
*   **Accent:** `#0000ff` (Azul Elétrico)
*   **Success/Positive:** `#008000` (Verde Técnico)
*   **Danger/Negative:** `#d00000` (Vermelho Alerta)

### 2.2 Tipografia
*   **Primary Font:** `Courier New` ou similar Mono (via Google Fonts).
*   **Estilo:** Tudo em caixa alta (Uppercase) para labels técnicos.
*   **Numeração:** Tabular nums (font-variant-numeric) para valores monetários.

### 2.3 Componentes Base
*   **BlueprintCard:** Container com fundo bege, borda de 2px preta e sombra rígida (offset 4px/4px black).
*   **TechnicalButton:** Botão retangular, borda 2px, sem arredondamento.
*   **SystemNavBar:** Bottom Navigation Bar com divisórias por bordas sólidas.

## 3. Arquitetura Técnica

### 3.1 Stack
*   **Framework:** Flutter 3.x
*   **Gerenciamento de Estado:** `flutter_riverpod`
*   **Navegação:** `go_router`
*   **Networking:** `dio` com interceptors para JWT (mesmo backend Go).
*   **Persistence:** `flutter_secure_storage` para tokens.

### 3.2 Estrutura de Pastas
```
lib/
├── core/
│   ├── theme/
│   ├── api/
│   └── router/
├── features/
│   ├── auth/
│   ├── dashboard/
│   ├── transactions/
│   ├── cards/
│   └── chat/
├── shared/
│   └── widgets/
└── main.dart
```

## 4. Especificação das Telas (Fase 1)

### 4.1 Dashboard (Híbrido)
*   **Topo:** Header técnico com "OPERADOR_ID" e timestamp.
*   **Visual:** Gráfico de Donut simplificado (CustomPainter ou fl_chart).
*   **Data Grid:** Lista de categorias em blocos 2x2 com barras de progresso (Telemetria).
*   **Logs:** Feed inferior com os últimos 3 eventos financeiros.

### 4.2 Transações (System Log)
*   Tabela técnica com headers invertidos (Black background, white text).
*   Filtros via "Handshake" (Bottom Sheet técnico).

### 4.3 Chat (Terminal Pierre)
*   Interface estilo console.
*   Input precedido por `>`.
*   Respostas do Pierre com prefixo `PIERRE_AI:`.

---

## 5. Próximos Passos
1. Criar o esqueleto do projeto Flutter.
2. Implementar o `BlueprintThemeData`.
3. Configurar a camada de serviço (Dio) integrada ao backend atual.
