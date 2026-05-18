import 'package:flutter/material.dart';
import '../theme/blueprint_theme.dart';

class BlueprintCard extends StatelessWidget {
  final Widget child;
  final String? label;
  final EdgeInsetsGeometry? padding;
  final double? height;

  const BlueprintCard({
    super.key,
    required this.child,
    this.label,
    this.padding,
    this.height,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        if (label != null) ...[
          Padding(
            padding: const EdgeInsets.only(left: 4, bottom: 6),
            child: Text(
              label!.toUpperCase(),
              style: const TextStyle(
                color: BlueprintTheme.textSecondary,
                fontWeight: FontWeight.bold,
                fontSize: 10,
                letterSpacing: 1.2,
              ),
            ),
          ),
        ],
        Container(
          height: height,
          padding: padding ?? const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: BlueprintTheme.surface,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: BlueprintTheme.border, width: 1),
          ),
          child: child,
        ),
      ],
    );
  }
}
