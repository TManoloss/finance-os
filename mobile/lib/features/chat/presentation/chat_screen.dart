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
                      hintText: 'DIGITE_SEU_COMANDO_AQUI...',
                      border: InputBorder.none,
                      hintStyle: TextStyle(fontWeight: FontWeight.bold, fontSize: 12),
                    ),
                    style: const TextStyle(fontWeight: FontWeight.bold),
                    onSubmitted: (_) => _sendMessage(),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.send),
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
