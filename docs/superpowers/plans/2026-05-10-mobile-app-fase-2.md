# Mobile App (Fase 2) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finalizar as features principais do App Flutter (Blueprint Edition), implementando o extrato de transações, o chat do Pierre e o roteamento via Bottom Navigation Bar.

**Architecture:** Riverpod para estado, GoRouter com `StatefulShellRoute` para a barra de navegação inferior persistente, e widgets altamente estilizados com o `BlueprintTheme`.

**Tech Stack:** Flutter, Riverpod, GoRouter, Dio.

---

### Task 1: Configuração do Bottom Navigation Bar (Shell Route)

**Files:**
- Modify: `mobile/lib/core/router/app_router.dart`
- Create: `mobile/lib/core/layout/main_layout.dart`

- [ ] **Step 1: Criar o layout base com a BottomNavigationBar (`mobile/lib/core/layout/main_layout.dart`)**

```dart
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons/lucide_icons.dart';

class MainLayout extends StatelessWidget {
  final StatefulNavigationShell navigationShell;

  const MainLayout({super.key, required this.navigationShell});

  void _goBranch(int index) {
    navigationShell.goBranch(
      index,
      initialLocation: index == navigationShell.currentIndex,
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: navigationShell,
      bottomNavigationBar: Container(
        decoration: BoxDecoration(
          border: const Border(top: BorderSide(color: Colors.black, width: 2)),
          color: const Color(0xFFE8E5DE),
        ),
        child: BottomNavigationBar(
          currentIndex: navigationShell.currentIndex,
          onTap: _goBranch,
          backgroundColor: Colors.transparent,
          elevation: 0,
          selectedItemColor: const Color(0xFF0000FF), // Accent Blue
          unselectedItemColor: Colors.black,
          showSelectedLabels: true,
          showUnselectedLabels: true,
          selectedLabelStyle: const TextStyle(fontWeight: FontWeight.w900, fontSize: 10),
          unselectedLabelStyle: const TextStyle(fontWeight: FontWeight.bold, fontSize: 10),
          type: BottomNavigationBarType.fixed,
          items: const [
            BottomNavigationBarItem(icon: Icon(LucideIcons.layoutDashboard), label: 'DASHBOARD'),
            BottomNavigationBarItem(icon: Icon(LucideIcons.arrowLeftRight), label: 'TRANSACTIONS'),
            BottomNavigationBarItem(icon: Icon(LucideIcons.creditCard), label: 'CREDIT'),
            BottomNavigationBarItem(icon: Icon(LucideIcons.terminal), label: 'PIERRE_AI'),
          ],
        ),
      ),
    );
  }
}
```

- [ ] **Step 2: Atualizar o GoRouter com `StatefulShellRoute` (`mobile/lib/core/router/app_router.dart`)**

```dart
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:finance_os/features/auth/presentation/login_screen.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_screen.dart';
import 'package:finance_os/core/layout/main_layout.dart';

final appRouter = GoRouter(
  initialLocation: '/login',
  routes: [
    GoRoute(
      path: '/login',
      name: 'login',
      builder: (context, state) => const LoginScreen(),
    ),
    StatefulShellRoute.indexedStack(
      builder: (context, state, navigationShell) {
        return MainLayout(navigationShell: navigationShell);
      },
      branches: [
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/dashboard',
              name: 'dashboard',
              builder: (context, state) => const DashboardScreen(),
            ),
          ],
        ),
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/transactions',
              name: 'transactions',
              builder: (context, state) => const Scaffold(body: Center(child: Text('TRANSACTIONS_MODULE'))),
            ),
          ],
        ),
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/cards',
              name: 'cards',
              builder: (context, state) => const Scaffold(body: Center(child: Text('CREDIT_MODULE'))),
            ),
          ],
        ),
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/chat',
              name: 'chat',
              builder: (context, state) => const Scaffold(body: Center(child: Text('PIERRE_AI_MODULE'))),
            ),
          ],
        ),
      ],
    ),
  ],
);
```

- [ ] **Step 3: Commit**

```bash
git add mobile/lib/core/router/app_router.dart mobile/lib/core/layout/main_layout.dart
git commit -m "feat: implement persistent bottom navigation bar"
```

---

### Task 2: Extrato de Transações (System Log)

**Files:**
- Create: `mobile/lib/features/transactions/data/transaction_model.dart`
- Create: `mobile/lib/features/transactions/presentation/transactions_provider.dart`
- Create: `mobile/lib/features/transactions/presentation/transactions_screen.dart`

- [ ] **Step 1: Criar o modelo `TransactionModel`**

```dart
// mobile/lib/features/transactions/data/transaction_model.dart
class TransactionModel {
  final String id;
  final String description;
  final double amount;
  final String direction;
  final String date;
  final String? categoryName;

  TransactionModel({
    required this.id,
    required this.description,
    required this.amount,
    required this.direction,
    required this.date,
    this.categoryName,
  });

  factory TransactionModel.fromJson(Map<String, dynamic> json) {
    return TransactionModel(
      id: json['id'] ?? '',
      description: json['description'] ?? '',
      amount: (json['amount'] as num).toDouble(),
      direction: json['direction'] ?? 'debit',
      date: json['date'] ?? '',
      categoryName: json['category'] != null ? json['category']['name'] : null,
    );
  }
}
```

- [ ] **Step 2: Criar o provider de transações**

```dart
// mobile/lib/features/transactions/presentation/transactions_provider.dart
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:finance_os/core/api/api_client.dart';
import 'package:finance_os/features/transactions/data/transaction_model.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';

final transactionsProvider = FutureProvider<List<TransactionModel>>((ref) async {
  final api = ref.watch(apiClientProvider);
  final resp = await api.dio.get('/transactions?page=1&page_size=50');
  final List<dynamic> data = resp.data['data']['transactions'] ?? [];
  return data.map((json) => TransactionModel.fromJson(json)).toList();
});
```

- [ ] **Step 3: Implementar a tela de transações com visual técnico**

```dart
// mobile/lib/features/transactions/presentation/transactions_screen.dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:finance_os/features/transactions/presentation/transactions_provider.dart';
import 'package:intl/intl.dart';

class TransactionsScreen extends ConsumerWidget {
  const TransactionsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final txsAsync = ref.watch(transactionsProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('SYSTEM_LOG // TRANSACTIONS')),
      body: txsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator(color: Colors.black)),
        error: (err, stack) => Center(child: Text('ERR_LOAD: $err')),
        data: (transactions) {
          if (transactions.isEmpty) return const Center(child: Text('BUFFER_EMPTY'));
          
          return ListView.separated(
            itemCount: transactions.length,
            separatorBuilder: (context, index) => const Divider(color: Colors.black, height: 2, thickness: 2),
            itemBuilder: (context, index) {
              final tx = transactions[index];
              final isCredit = tx.direction == 'credit';
              final color = isCredit ? const Color(0xFF008000) : const Color(0xFFD00000);
              final sign = isCredit ? '+' : '-';
              
              // Format date from "2026-05-10T00:00:00Z" to "10/05"
              String formattedDate = tx.date;
              try {
                final dt = DateTime.parse(tx.date);
                formattedDate = DateFormat('dd/MM').format(dt);
              } catch (_) {}

              return Container(
                padding: const EdgeInsets.all(16),
                color: const Color(0xFFF4F1EA),
                child: Row(
                  children: [
                    Container(
                      width: 40,
                      height: 40,
                      decoration: BoxDecoration(
                        border: Border.all(color: Colors.black, width: 2),
                        color: isCredit ? color : Colors.transparent,
                      ),
                      alignment: Alignment.center,
                      child: Text(
                        sign,
                        style: TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.bold,
                          color: isCredit ? Colors.white : color,
                        ),
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(tx.description, style: const TextStyle(fontWeight: FontWeight.w900), maxLines: 1, overflow: TextOverflow.ellipsis),
                          const SizedBox(height: 4),
                          Text('DATE: $formattedDate | CAT: ${tx.categoryName ?? "NULL"}', style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold)),
                        ],
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$').format(tx.amount),
                      style: TextStyle(fontWeight: FontWeight.w900, fontSize: 16, color: color),
                    ),
                  ],
                ),
              );
            },
          );
        },
      ),
    );
  }
}
```

- [ ] **Step 4: Atualizar a rota no GoRouter**
Mude `builder: (context, state) => const Scaffold(...)` na rota `/transactions` (em `mobile/lib/core/router/app_router.dart`) para:
`builder: (context, state) => const TransactionsScreen(),`
(Adicione o import correspondente).

- [ ] **Step 5: Commit**

```bash
git add mobile/lib/features/transactions/ mobile/lib/core/router/app_router.dart
git commit -m "feat: implement transactions system log screen"
```

---

### Task 3: Chat do Pierre (Terminal)

**Files:**
- Create: `mobile/lib/features/chat/presentation/chat_screen.dart`
- Modify: `mobile/lib/core/router/app_router.dart`

- [ ] **Step 1: Implementar a interface estilo console**

```dart
// mobile/lib/features/chat/presentation/chat_screen.dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:finance_os/core/api/api_client.dart';
import 'package:finance_os/features/dashboard/presentation/dashboard_provider.dart';

class ChatMessage {
  final String role;
  final String content;
  ChatMessage(this.role, this.content);
}

class ChatScreen extends ConsumerStatefulWidget {
  const ChatScreen({super.key});
  @override
  ConsumerState<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends ConsumerState<ChatScreen> {
  final TextEditingController _inputController = TextEditingController();
  final List<ChatMessage> _messages = [
    ChatMessage('assistant', 'SISTEMA_INICIALIZADO: Olá! Eu sou o Pierre. Como posso auxiliar na sua análise financeira hoje?')
  ];
  bool _isLoading = false;

  Future<void> _sendMessage() async {
    final text = _inputController.text.trim();
    if (text.isEmpty || _isLoading) return;

    setState(() {
      _messages.add(ChatMessage('user', text));
      _inputController.clear();
      _isLoading = true;
    });

    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.dio.post('/chat', data: {
        'message': text,
        'history': _messages.map((m) => {'role': m.role, 'content': m.content}).toList(),
      });
      
      setState(() {
        _messages.add(ChatMessage('assistant', resp.data['data']['response']));
      });
    } catch (e) {
      setState(() {
        _messages.add(ChatMessage('assistant', 'ERROR_REF_0xChat: Falha na comunicação com o núcleo de IA.'));
      });
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('TERMINAL_PIERRE')),
      body: Column(
        children: [
          Expanded(
            child: Container(
              color: const Color(0xFFF4F1EA),
              child: ListView.builder(
                padding: const EdgeInsets.all(16),
                itemCount: _messages.length + (_isLoading ? 1 : 0),
                itemBuilder: (context, index) {
                  if (index == _messages.length) {
                    return const Padding(
                      padding: EdgeInsets.only(top: 16),
                      child: Text('> PROCESSANDO_REQUISICAO...', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
                    );
                  }
                  
                  final msg = _messages[index];
                  final isUser = msg.role == 'user';
                  
                  return Container(
                    margin: const EdgeInsets.only(bottom: 16),
                    child: Column(
                      crossAxisAlignment: isUser ? CrossAxisAlignment.end : CrossAxisAlignment.start,
                      children: [
                        Text(
                          isUser ? 'USUARIO' : 'PIERRE_AI', 
                          style: const TextStyle(fontSize: 10, fontWeight: FontWeight.w900)
                        ),
                        const SizedBox(height: 4),
                        Container(
                          padding: const EdgeInsets.all(16),
                          decoration: BoxDecoration(
                            color: isUser ? const Color(0xFF0000FF) : const Color(0xFFE8E5DE),
                            border: Border.all(color: Colors.black, width: 2),
                            boxShadow: const [BoxShadow(color: Colors.black, offset: Offset(4, 4))],
                          ),
                          child: Text(
                            msg.content,
                            style: TextStyle(
                              color: isUser ? Colors.white : Colors.black,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ],
                    ),
                  );
                },
              ),
            ),
          ),
          Container(
            padding: const EdgeInsets.all(16),
            decoration: const BoxDecoration(
              color: Color(0xFFE8E5DE),
              border: Border(top: BorderSide(color: Colors.black, width: 2)),
            ),
            child: Row(
              children: [
                const Text('>', style: TextStyle(fontWeight: FontWeight.w900, fontSize: 20)),
                const SizedBox(width: 8),
                Expanded(
                  child: TextField(
                    controller: _inputController,
                    decoration: const InputDecoration(
                      hintText: 'DIGITE_SUA_COMANDO_AQUI...',
                      border: InputBorder.none,
                      hintStyle: TextStyle(fontWeight: FontWeight.bold, fontSize: 12),
                    ),
                    style: const TextStyle(fontWeight: FontWeight.bold),
                    onSubmitted: (_) => _sendMessage(),
                  ),
                ),
                IconButton(
                  icon: const Icon(LucideIcons.send),
                  onPressed: _isLoading ? null : _sendMessage,
                  color: Colors.black,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
```

- [ ] **Step 2: Atualizar a rota no GoRouter**
Mude `builder: (context, state) => const Scaffold(...)` na rota `/chat` (em `mobile/lib/core/router/app_router.dart`) para:
`builder: (context, state) => const ChatScreen(),`
(Adicione o import correspondente).

- [ ] **Step 3: Commit**

```bash
git add mobile/lib/features/chat/ mobile/lib/core/router/app_router.dart
git commit -m "feat: implement pierre terminal chat screen"
```
