import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_markdown/flutter_markdown.dart';
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
        title: const Text(
          'TERMINAL_PIERRE_V1.1',
          style: TextStyle(fontWeight: FontWeight.bold, letterSpacing: 1),
        ),
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: false,
      ),
      backgroundColor: const Color(0xFFD4D1CA),
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
                    padding: EdgeInsets.only(top: 16),
                    child: Text('> PROCESSANDO_REQUISICAO...', 
                      style: TextStyle(fontWeight: FontWeight.bold, fontSize: 10, color: Colors.blueGrey)),
                  );
                }
                
                final msg = _messages[index];
                final isUser = msg.role == 'user';
                
                return Container(
                  margin: const EdgeInsets.only(bottom: 24),
                  child: Column(
                    crossAxisAlignment: isUser ? CrossAxisAlignment.end : CrossAxisAlignment.start,
                    children: [
                      Text(
                        isUser ? 'USUARIO' : 'PIERRE_AI', 
                        style: const TextStyle(fontSize: 10, fontWeight: FontWeight.w900, letterSpacing: 1)
                      ),
                      const SizedBox(height: 6),
                      Container(
                        padding: const EdgeInsets.all(16),
                        constraints: BoxConstraints(
                          maxWidth: MediaQuery.of(context).size.width * 0.85,
                        ),
                        decoration: BoxDecoration(
                          color: isUser ? const Color(0xFF2A2A2A) : Colors.white,
                          border: Border.all(color: Colors.black, width: 2),
                          boxShadow: const [BoxShadow(color: Colors.black, offset: Offset(4, 4))],
                        ),
                        child: MarkdownBody(
                          data: msg.content,
                          styleSheet: MarkdownStyleSheet(
                            p: TextStyle(
                              color: isUser ? Colors.white : Colors.black,
                              fontWeight: FontWeight.bold,
                              fontSize: 14,
                            ),
                            h1: TextStyle(
                              color: isUser ? Colors.white : Colors.black,
                              fontWeight: FontWeight.w900,
                              fontSize: 18,
                            ),
                            h2: TextStyle(
                              color: isUser ? Colors.white : Colors.black,
                              fontWeight: FontWeight.w900,
                              fontSize: 16,
                            ),
                            listBullet: TextStyle(
                              color: isUser ? Colors.white : Colors.black,
                              fontWeight: FontWeight.bold,
                            ),
                            code: TextStyle(
                              backgroundColor: isUser ? Colors.white12 : Colors.black12,
                              fontFamily: 'monospace',
                              fontWeight: FontWeight.bold,
                            ),
                            blockquote: const TextStyle(
                              fontStyle: FontStyle.italic,
                              color: Colors.blueGrey,
                            ),
                            blockquoteDecoration: BoxDecoration(
                              border: const Border(left: BorderSide(color: Colors.black, width: 4)),
                              color: isUser ? Colors.white12 : Colors.black12,
                            ),
                          ),
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
              color: Colors.white,
              border: Border(top: BorderSide(color: Colors.black, width: 2)),
            ),
            child: Row(
              children: [
                const Text('>', style: TextStyle(fontWeight: FontWeight.w900, fontSize: 20)),
                const SizedBox(width: 12),
                Expanded(
                  child: TextField(
                    controller: _inputController,
                    decoration: const InputDecoration(
                      hintText: 'DIGITE_SEU_COMANDO_AQUI...',
                      border: InputBorder.none,
                      hintStyle: TextStyle(fontWeight: FontWeight.bold, fontSize: 12, color: Colors.grey),
                    ),
                    style: const TextStyle(fontWeight: FontWeight.bold),
                    onSubmitted: (_) => _sendMessage(),
                  ),
                ),
                Container(
                  decoration: const BoxDecoration(
                    color: Colors.black,
                    boxShadow: [BoxShadow(color: Colors.grey, offset: Offset(2, 2))],
                  ),
                  child: IconButton(
                    icon: const Icon(Icons.send, color: Colors.white),
                    onPressed: _isLoading ? null : _sendMessage,
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
