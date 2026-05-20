import 'package:flutter/material.dart';
import 'package:finance_os/core/theme/blueprint_theme.dart';

class ActivityFeedWidget extends StatelessWidget {
  final List<dynamic> feedItems;
  const ActivityFeedWidget({super.key, required this.feedItems});

  @override
  Widget build(BuildContext context) {
    final items = feedItems.take(5).toList();

    if (items.isEmpty) {
      return const SizedBox.shrink();
    }

    return Container(
      decoration: BoxDecoration(
        color: BlueprintTheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: BlueprintTheme.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 16, 16, 8),
            child: Text('FEED_DE_ATIVIDADES', style: TextStyle(fontSize: 10, fontWeight: FontWeight.w900, letterSpacing: 1, color: BlueprintTheme.accentPurple)),
          ),
          ...items.map((item) {
            final type = (item['type'] ?? '').toString();
            final message = (item['message'] ?? '').toString();
            final severity = (item['severity'] ?? 'low').toString();

            Color dotColor;
            switch (severity) {
              case 'high': dotColor = BlueprintTheme.danger; break;
              case 'medium': dotColor = BlueprintTheme.warning; break;
              default: dotColor = BlueprintTheme.accentTeal;
            }

            IconData icon;
            switch (type) {
              case 'alert': icon = Icons.warning_amber_rounded; break;
              case 'achievement': icon = Icons.emoji_events_rounded; break;
              case 'sync': icon = Icons.sync_rounded; break;
              case 'insight': icon = Icons.lightbulb_outline_rounded; break;
              default: icon = Icons.info_outline_rounded;
            }

            return Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    width: 28, height: 28,
                    decoration: BoxDecoration(
                      color: dotColor.withValues(alpha: 0.15),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Icon(icon, color: dotColor, size: 14),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      message,
                      style: const TextStyle(fontSize: 12, height: 1.4, color: BlueprintTheme.textPrimary),
                    ),
                  ),
                ],
              ),
            );
          }),
          const SizedBox(height: 8),
        ],
      ),
    );
  }
}
