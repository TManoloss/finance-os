import 'package:flutter/material.dart';
import '../../../shared/widgets/blueprint_card.dart';

class DashboardScreen extends StatelessWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFD4D1CA),
      appBar: AppBar(
        title: const Text(
          'DASHBOARD_V1.0',
          style: TextStyle(
            fontWeight: FontWeight.bold,
            letterSpacing: 2,
          ),
        ),
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: false,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const BlueprintCard(
              label: 'TELEMETRIA_GASTOS',
              height: 200,
              child: Center(
                child: Text(
                  '[GRAFICO_DONUT_AQUI]',
                  style: TextStyle(
                    fontFamily: 'monospace',
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ),
            const SizedBox(height: 24),
            GridView.count(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              crossAxisCount: 2,
              crossAxisSpacing: 16,
              mainAxisSpacing: 16,
              childAspectRatio: 1.2,
              children: const [
                BlueprintCard(
                  label: 'ALIMENTACAO',
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        'R\$ 1.250,00',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      Text(
                        'REF: MAIO',
                        style: TextStyle(fontSize: 10),
                      ),
                    ],
                  ),
                ),
                BlueprintCard(
                  label: 'TRANSPORTE',
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        'R\$ 450,00',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      Text(
                        'REF: MAIO',
                        style: TextStyle(fontSize: 10),
                      ),
                    ],
                  ),
                ),
                BlueprintCard(
                  label: 'SAUDE',
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        'R\$ 320,00',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      Text(
                        'REF: MAIO',
                        style: TextStyle(fontSize: 10),
                      ),
                    ],
                  ),
                ),
                BlueprintCard(
                  label: 'LAZER',
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        'R\$ 890,00',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      Text(
                        'REF: MAIO',
                        style: TextStyle(fontSize: 10),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
