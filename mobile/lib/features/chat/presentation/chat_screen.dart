import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_markdown/flutter_markdown.dart';
import '../../theme/blueprint_theme.dart';
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
      
      setState(() {
        _messages.add(ChatMessage('assistant', resp.data['data']['response']));
      });
      _scrollToBottom();
    } catch (e) {
      setState(() {
        _messages.add(ChatMessage('assistant', 'ERROR_REF_0xChat: Falha na comunicação com o núcleo de IA.'));
      });
      _scrollToBottom();
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('TERMINAL PIERRE'),
      ),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              controller: _scrollController,
              padding: const EdgeInsets.all(16),
              itemCount: _messages.length + (_isLoading ? 1 : 0),
              itemBuilder: (context, index) {
                if (index == _messages.length) {
                  return const Padding(
                    padding: EdgeInsets.symmetric(vertical: 16),
                    child: Text(
                      '> PROCESSANDO_REQUISICAO...', 
                      style: TextStyle(
                        fontWeight: FontWeight.bold, 
                        fontSize: 10, 
                        color: BlueprintTheme.accentPurple
                      )
                    ),
                  );
                }
                
                final msg = _messages[index];
                final isUser = msg.role == 'user';
                
                return Container(
                  margin: const EdgeInsets.only(bottom: 24),
                  child: Row(
                    mainAxisAlignment: isUser ? MainAxisAlignment.end : MainAxisAlignment.start,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      if (!isUser) ...[
                        Container(
                          width: 24,
                          height: 24,
                          margin: const EdgeInsets.only(top: 4, right: 12),
                          decoration: BoxDecoration(
                            color: BlueprintTheme.accentPurple.withOpacity(0.2),
                            borderRadius: BorderRadius.circular(6),
                            border: Border.all(color: BlueprintTheme.accentPurple, width: 1),
                          ),
                          child: const Center(
                            child: Text(
                              'P', 
                              style: TextStyle(
                                color: BlueprintTheme.accentPurple, 
                                fontSize: 12, 
                                fontWeight: FontWeight.bold
                              )
                            ),
                          ),
                        ),
                      ],
                      Flexible(
                        child: Column(
                          crossAxisAlignment: isUser ? CrossAxisAlignment.end : MainAxisAlignment.start,
                          children: [
                            Text(
                              isUser ? 'VOCÊ' : 'PIERRE_AI', 
                              style: const TextStyle(
                                fontSize: 8, 
                                fontWeight: FontWeight.bold, 
                                color: BlueprintTheme.textSecondary,
                                letterSpacing: 1
                              )
                            ),
                            const SizedBox(height: 6),
                            Container(
                              padding: const EdgeInsets.all(16),
                              decoration: BoxDecoration(
                                color: isUser ? BlueprintTheme.accentPurple : BlueprintTheme.surface,
                                borderRadius: BorderRadius.circular(16).copyWith(
                                  topRight: isUser ? const Radius.circular(4) : const Radius.circular(16),
                                  topLeft: isUser ? const Radius.circular(16) : const Radius.circular(4),
                                ),
                                border: Border.all(
                                  color: isUser ? BlueprintTheme.accentPurple : BlueprintTheme.border, 
                                  width: 1
                                ),
                              ),
                              child: MarkdownBody(
                                data: msg.content,
                                styleSheet: MarkdownStyleSheet(
                                  p: TextStyle(
                                    color: isUser ? Colors.white : BlueprintTheme.textPrimary,
                                    fontSize: 14,
                                    height: 1.5,
                                  ),
                                  strong: const TextStyle(fontWeight: FontWeight.bold),
                                  code: TextStyle(
                                    backgroundColor: Colors.black26,
                                    fontFamily: 'monospace',
                                    fontSize: 12,
                                    color: isUser ? Colors.white : BlueprintTheme.accentTeal,
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
          Container(
            padding: EdgeInsets.only(
              left: 16, 
              right: 16, 
              top: 16, 
              bottom: MediaQuery.of(context).padding.bottom + 16
            ),
            decoration: const BoxDecoration(
              color: BlueprintTheme.surface,
              border: Border(top: BorderSide(color: BlueprintTheme.border, width: 1)),
            ),
            child: Row(
              children: [
                Expanded(
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    decoration: BoxDecoration(
                      color: BlueprintTheme.elevated,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: BlueprintTheme.border, width: 1),
                    ),
                    child: TextField(
                      controller: _inputController,
                      decoration: const InputDecoration(
                        hintText: 'Pergunte ao Pierre...',
                        border: InputBorder.none,
                        hintStyle: TextStyle(fontSize: 12, color: BlueprintTheme.textSecondary),
                      ),
                      style: const TextStyle(fontSize: 14),
                      onSubmitted: (_) => _sendMessage(),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                GestureDetector(
                  onTap: _isLoading ? null : _sendMessage,
                  child: Container(
                    width: 48,
                    height: 48,
                    decoration: BoxDecoration(
                      color: BlueprintTheme.accentPurple,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Icon(Icons.send_rounded, color: Colors.white, size: 20),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
