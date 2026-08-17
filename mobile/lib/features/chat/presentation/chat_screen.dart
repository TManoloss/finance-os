import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_markdown/flutter_markdown.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';
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
  final ScrollController _scrollController = ScrollController();
  final List<ChatMessage> _messages = [
    ChatMessage('assistant', 'SISTEMA_INICIALIZADO: Olá! Eu sou o Pierre. Como posso auxiliar na sua análise financeira hoje?')
  ];
  bool _isLoading = false;

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollController.hasClients) {
        _scrollController.animateTo(
          _scrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeOut,
        );
      }
    });
  }

  Future<void> _sendMessage() async {
    final text = _inputController.text.trim();
    if (text.isEmpty || _isLoading) return;
    setState(() {
      _messages.add(ChatMessage('user', text));
      _inputController.clear();
      _isLoading = true;
    });
    _scrollToBottom();
    try {
      final api = ref.read(apiClientProvider);
      final resp = await api.dio.post('/chat', data: {
        'message': text,
        'history': _messages.map((m) => {'role': m.role, 'content': m.content}).toList(),
      });
      setState(() => _messages.add(ChatMessage('assistant', resp.data['data']['response'])));
    } catch (_) {
      setState(() => _messages.add(ChatMessage('assistant', 'ERROR_REF_0xChat: Falha na comunicação com o núcleo de IA.')));
    } finally {
      setState(() => _isLoading = false);
      _scrollToBottom();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: BlueprintTheme.background,
      appBar: AppBar(
        title: const Text('Pierre'),
      ),
      body: Column(
        children: [
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            margin: const EdgeInsets.fromLTRB(16, 8, 16, 0),
            decoration: BoxDecoration(color: BlueprintTheme.elevated, borderRadius: BorderRadius.circular(12)),
            child: Row(children: [
              Container(width: 6, height: 6, color: BlueprintTheme.accentTeal),
              const SizedBox(width: 6),
              Text('Pierre está online', style: terminalLabel(fontSize: 11)),
            ]),
          ),

          // Mensagens
          Expanded(
            child: ListView.builder(
              controller: _scrollController,
              padding: const EdgeInsets.all(16),
              itemCount: _messages.length + (_isLoading ? 1 : 0),
              itemBuilder: (_, index) {
                if (index == _messages.length) {
                  return Padding(
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    child: Text('> PROCESSANDO_REQUISICAO...', style: terminalLabel(color: BlueprintTheme.accentPurple, fontSize: 10)),
                  );
                }
                final msg = _messages[index];
                final isUser = msg.role == 'user';
                return Container(
                  margin: const EdgeInsets.only(bottom: 20),
                  child: Row(
                    mainAxisAlignment: isUser ? MainAxisAlignment.end : MainAxisAlignment.start,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      if (!isUser) ...[
                        Padding(
                          padding: const EdgeInsets.only(top: 2, right: 10),
                          child: CircleAvatar(
                          radius: 12,
                          backgroundColor: BlueprintTheme.accentPurple,
                          child: const Center(child: Text('P', style: TextStyle(color: Colors.white, fontFamily: 'monospace', fontWeight: FontWeight.w900, fontSize: 12))),
                        )),
                      ],
                      Flexible(
                        child: Column(
                          crossAxisAlignment: isUser ? CrossAxisAlignment.end : CrossAxisAlignment.start,
                          children: [
                            Text(isUser ? 'Você' : 'Pierre', style: terminalLabel(fontSize: 10)),
                            const SizedBox(height: 4),
                            Container(
                              padding: const EdgeInsets.all(14),
                              decoration: BoxDecoration(
                                color: isUser ? BlueprintTheme.accentPurple : BlueprintTheme.surface,
                                borderRadius: BorderRadius.circular(16),
                              ),
                              child: MarkdownBody(
                                data: msg.content,
                                styleSheet: MarkdownStyleSheet(
                                  p: TextStyle(
                                    color: isUser ? BlueprintTheme.surface : BlueprintTheme.textPrimary,
                                    fontSize: 14, height: 1.5,
                                  ),
                                  strong: TextStyle(fontWeight: FontWeight.w900, color: isUser ? BlueprintTheme.surface : BlueprintTheme.textPrimary),
                                  code: TextStyle(
                                    backgroundColor: isUser ? Colors.white12 : BlueprintTheme.elevated,
                                    fontFamily: 'monospace', fontSize: 12,
                                    color: isUser ? Colors.white : BlueprintTheme.accentPurple,
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                );
              },
            ),
          ),

          // Input neo-brutal
          Container(
            padding: EdgeInsets.only(
              left: 16, right: 16, top: 12,
              bottom: MediaQuery.of(context).padding.bottom + 12,
            ),
            decoration: const BoxDecoration(color: BlueprintTheme.background),
            child: Row(children: [
              Expanded(
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  decoration: BoxDecoration(color: BlueprintTheme.elevated, borderRadius: BorderRadius.circular(14)),
                  child: TextField(
                    controller: _inputController,
                    decoration: InputDecoration(
                      hintText: 'Pergunte ao Pierre...',
                      border: InputBorder.none,
                      hintStyle: terminalLabel(fontSize: 12),
                    ),
                    style: const TextStyle(fontSize: 14),
                    onSubmitted: (_) => _sendMessage(),
                  ),
                ),
              ),
              const SizedBox(width: 10),
              GestureDetector(
                onTap: _isLoading ? null : _sendMessage,
                child: Container(
                  width: 48, height: 48,
                  decoration: BoxDecoration(color: BlueprintTheme.accentPurple, borderRadius: BorderRadius.circular(14)),
                  child: const Icon(LucideIcons.send, color: Colors.white, size: 18),
                ),
              ),
            ]),
          ),
        ],
      ),
    );
  }
}
