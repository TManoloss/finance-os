import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

// Este é um mock para demonstração do conceito
// Em um cenário real, os dados viriam do Riverpod Provider
class ReplayScreen extends ConsumerStatefulWidget {
  final String month;

  const ReplayScreen({Key? key, required this.month}) : super(key: key);

  @override
  ConsumerState<ReplayScreen> createState() => _ReplayScreenState();
}

class _ReplayScreenState extends ConsumerState<ReplayScreen> {
  int _currentSlide = 0;
  
  // Mock data representing the narrative JSON from the agent
  final List<Map<String, dynamic>> _slides = [
    {
      "title": "Abril foi agitado!",
      "text": "Você focou bastante em Alimentação, que representou grande parte do seu orçamento.",
      "highlight_value": "R\$ 1.230"
    },
    {
      "title": "Aquele Ifood salvou...",
      "text": "Seu merchant favorito foi o iFood, com 12 pedidos este mês.",
      "highlight_value": "Top 1"
    },
    {
      "title": "Maior Surpresa",
      "text": "Aquela compra no Mercado Livre foi a maior do mês.",
      "highlight_value": "R\$ 450"
    }
  ];

  void _nextSlide() {
    if (_currentSlide < _slides.length - 1) {
      setState(() {
        _currentSlide++;
      });
    } else {
      context.pop();
    }
  }

  void _prevSlide() {
    if (_currentSlide > 0) {
      setState(() {
        _currentSlide--;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final slide = _slides[_currentSlide];

    return Scaffold(
      backgroundColor: BlueprintTheme.background,
      body: SafeArea(
        child: Stack(
          children: [
            // Progress Indicators
            Positioned(
              top: 16,
              left: 16,
              right: 16,
              child: Row(
                children: List.generate(_slides.length, (index) {
                  return Expanded(
                    child: Container(
                      margin: const EdgeInsets.symmetric(horizontal: 4),
                      height: 4,
                      decoration: BoxDecoration(
                        color: index <= _currentSlide
                            ? BlueprintTheme.accentPurple
                            : BlueprintTheme.border,
                        borderRadius: BorderRadius.circular(2),
                      ),
                    ),
                  );
                }),
              ),
            ),
            
            // Gestures
            Positioned.fill(
              child: Row(
                children: [
                  Expanded(
                    flex: 1,
                    child: GestureDetector(
                      onTap: _prevSlide,
                      behavior: HitTestBehavior.opaque,
                    ),
                  ),
                  Expanded(
                    flex: 2,
                    child: GestureDetector(
                      onTap: _nextSlide,
                      behavior: HitTestBehavior.opaque,
                    ),
                  ),
                ],
              ),
            ),

            // Content
            Center(
              child: Padding(
                padding: const EdgeInsets.all(32.0),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      slide['title'],
                      style: const TextStyle(
                        color: BlueprintTheme.accentTeal,
                        fontSize: 28,
                        fontWeight: FontWeight.bold,
                      ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 24),
                    Text(
                      slide['text'],
                      style: const TextStyle(
                        color: BlueprintTheme.textPrimary,
                        fontSize: 20,
                        height: 1.5,
                      ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 48),
                    if (slide['highlight_value'] != null)
                      Text(
                        slide['highlight_value'],
                        style: const TextStyle(
                          color: BlueprintTheme.accentPurple,
                          fontSize: 48,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                  ],
                ),
              ),
            ),

            // Close button
            Positioned(
              bottom: 32,
              left: 0,
              right: 0,
              child: Center(
                child: IconButton(
                  icon: const Icon(Icons.close, color: BlueprintTheme.textSecondary),
                  onPressed: () => context.pop(),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
