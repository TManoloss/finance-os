# Mobile App (Fase 1) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Criar o esqueleto funcional do App Flutter com a estética Blueprint, incluindo autenticação e o dashboard híbrido.

**Architecture:** Clean Architecture simplificada, utilizando Riverpod para gerenciamento de estado, GoRouter para navegação e Dio para chamadas de API. O design será centralizado em um BlueprintThemeData customizado.

**Tech Stack:** Flutter, Riverpod, GoRouter, Dio, Google Fonts (Courier Prime).

---

### Task 1: Scaffolding e Dependências

**Files:**
- Create: `mobile/pubspec.yaml`
- Create: `mobile/lib/main.dart`

- [ ] **Step 1: Criar o arquivo pubspec.yaml com as dependências necessárias**

```yaml
name: finance_os
description: Personal Finance OS Mobile - Blueprint Edition
version: 1.0.0+1
publish_to: 'none'

environment:
  sdk: '>=3.0.0 <4.0.0'

dependencies:
  flutter:
    sdk: flutter
  flutter_riverpod: ^2.5.1
  go_router: ^13.2.0
  dio: ^5.4.1
  flutter_secure_storage: ^9.0.0
  google_fonts: ^6.2.1
  lucide_icons: ^0.321.0
  intl: ^0.19.0
  fl_chart: ^0.66.0

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^3.0.0

flutter:
  uses-material-design: true
```

- [ ] **Step 2: Criar o ponto de entrada main.dart**

```dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';

void main() {
  runApp(
    const ProviderScope(
      child: FinanceOSApp(),
    ),
  );
}

class FinanceOSApp extends StatelessWidget {
  const FinanceOSApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'FINANCE_OS',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        useMaterial3: true,
        textTheme: GoogleFonts.courierPrimeTextTheme(),
      ),
      home: const Scaffold(
        body: Center(child: Text('FINANCE_OS_BOOT...')),
      ),
    );
  }
}
```

- [ ] **Step 3: Commit inicial**

```bash
git add mobile/pubspec.yaml mobile/lib/main.dart
git commit -m "chore: initial mobile scaffold"
```

---

### Task 2: Implementação do BlueprintTheme

**Files:**
- Create: `mobile/lib/core/theme/blueprint_theme.dart`

- [ ] **Step 1: Definir as constantes de cores e o ThemeData customizado**

```dart
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class BlueprintTheme {
  static const Color background = Color(0xFFF4F1EA);
  static const Color border = Color(0xFF000000);
  static const Color elevated = Color(0xFFE8E5DE);
  static const Color accent = Color(0xFF0000FF);
  static const Color danger = Color(0xFFD00000);
  static const Color success = Color(0xFF008000);

  static ThemeData get light {
    return ThemeData(
      useMaterial3: true,
      scaffoldBackgroundColor: background,
      colorScheme: ColorScheme.fromSeed(
        seedColor: accent,
        background: background,
        surface: background,
        onSurface: border,
        primary: accent,
        error: danger,
      ),
      textTheme: GoogleFonts.courierPrimeTextTheme().apply(
        bodyColor: border,
        displayColor: border,
      ),
      appBarTheme: const AppBarTheme(
        backgroundColor: elevated,
        elevation: 0,
        shape: Border(bottom: BorderSide(color: border, width: 2)),
        titleTextStyle: TextStyle(color: border, fontWeight: FontWeight.bold, fontSize: 18),
      ),
      cardTheme: CardTheme(
        color: background,
        elevation: 0,
        shape: RoundedRectangleBorder(
          side: const BorderSide(color: border, width: 2),
          borderRadius: BorderRadius.zero,
        ),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: border,
          foregroundColor: Colors.white,
          shape: const RoundedRectangleBorder(borderRadius: BorderRadius.zero),
          padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 24),
          textStyle: const TextStyle(fontWeight: FontWeight.bold, letterSpacing: 1.2),
        ),
      ),
    );
  }
}
```

- [ ] **Step 2: Atualizar main.dart para usar o novo tema**

```dart
// Modificar mobile/lib/main.dart
import 'package:finance_os/core/theme/blueprint_theme.dart';

// ... no build do FinanceOSApp ...
return MaterialApp(
  // ...
  theme: BlueprintTheme.light,
  // ...
);
```

- [ ] **Step 3: Commit**

```bash
git add mobile/lib/core/theme/blueprint_theme.dart mobile/lib/main.dart
git commit -m "feat: add blueprint theme implementation"
```

---

### Task 3: Camada de API e Autenticação

**Files:**
- Create: `mobile/lib/core/api/api_client.dart`
- Create: `mobile/lib/features/auth/data/auth_repository.dart`

- [ ] **Step 1: Criar o cliente Dio com Interceptor**

```dart
import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class ApiClient {
  final Dio dio;
  final storage = const FlutterSecureStorage();

  ApiClient() : dio = Dio(BaseOptions(baseUrl: 'http://localhost:8080/api/v1')) {
    dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await storage.read(key: 'access_token');
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        return handler.next(options);
      },
    ));
  }
}
```

- [ ] **Step 2: Criar o repositório de autenticação**

```dart
import 'package:finance_os/core/api/api_client.dart';

class AuthRepository {
  final ApiClient api;
  AuthRepository(this.api);

  Future<bool> login(String email, String password) async {
    try {
      final resp = await api.dio.post('/auth/login', data: {
        'email': email,
        'password': password,
      });
      if (resp.statusCode == 200) {
        final data = resp.data['data'];
        await api.storage.write(key: 'access_token', value: data['access_token']);
        return true;
      }
      return false;
    } catch (e) {
      return false;
    }
  }
}
```

- [ ] **Step 3: Commit**

```bash
git add mobile/lib/core/api/api_client.dart mobile/lib/features/auth/data/auth_repository.dart
git commit -m "feat: add api client and auth repository"
```

---

### Task 4: Tela de Login (Blueprint Terminal)

**Files:**
- Create: `mobile/lib/features/auth/presentation/login_screen.dart`

- [ ] **Step 1: Implementar o visual de terminal na tela de login**

```dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});
  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Text('FINANCE_OS', style: TextStyle(fontSize: 32, fontWeight: FontWeight.w900)),
              const Text('CORE_ENGINE_LOGIN_REQUIRED', style: TextStyle(fontSize: 10)),
              const SizedBox(height: 40),
              TextField(
                controller: _emailController,
                decoration: const InputDecoration(
                  labelText: 'INPUT_IDENTITY',
                  border: OutlineInputBorder(borderSide: BorderSide(width: 2)),
                ),
              ),
              const SizedBox(height: 20),
              TextField(
                controller: _passwordController,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'SECURE_CREDENTIAL',
                  border: OutlineInputBorder(borderSide: BorderSide(width: 2)),
                ),
              ),
              const SizedBox(height: 40),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () {}, // Lógica de login aqui
                  child: const Text('EXECUTE_LOGIN'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add mobile/lib/features/auth/presentation/login_screen.dart
git commit -m "feat: add blueprint login screen"
```

---

### Task 5: Dashboard Híbrido

**Files:**
- Create: `mobile/lib/features/dashboard/presentation/dashboard_screen.dart`
- Create: `mobile/lib/shared/widgets/blueprint_card.dart`

- [ ] **Step 1: Criar o widget BlueprintCard**

```dart
import 'package:flutter/material.dart';

class BlueprintCard extends StatelessWidget {
  final Widget child;
  final String? label;
  const BlueprintCard({super.key, required this.child, this.label});

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (label != null) Padding(
          padding: const EdgeInsets.only(bottom: 8),
          child: Text(label!, style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold)),
        ),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: const Color(0xFFF4F1EA),
            border: Border.all(color: Colors.black, width: 2),
            boxShadow: const [BoxShadow(color: Colors.black, offset: Offset(4, 4))],
          ),
          child: child,
        ),
      ],
    );
  }
}
```

- [ ] **Step 2: Implementar o Dashboard com o layout híbrido (Gráfico + Grid)**

```dart
import 'package:flutter/material.dart';
import 'package:finance_os/shared/widgets/blueprint_card.dart';

class DashboardScreen extends StatelessWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('DASHBOARD_V1.0')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          const BlueprintCard(
            label: 'TELEMETRIA_DE_GASTOS',
            child: SizedBox(
              height: 200,
              child: Center(child: Text('[GRAFICO_DONUT_AQUI]')),
            ),
          ),
          const SizedBox(height: 20),
          GridView.count(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            crossAxisCount: 2,
            mainAxisSpacing: 16,
            crossAxisSpacing: 16,
            children: const [
              BlueprintCard(child: Text('ALIMENT.\nR\$ 1.832')),
              BlueprintCard(child: Text('TRANSP.\nR\$ 450')),
            ],
          ),
        ],
      ),
    );
  }
}
```

- [ ] **Step 3: Commit**

```bash
git add mobile/lib/features/dashboard/presentation/dashboard_screen.dart mobile/lib/shared/widgets/blueprint_card.dart
git commit -m "feat: add hybrid dashboard screen"
```
