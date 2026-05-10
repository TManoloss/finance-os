import 'package:flutter/material.dart';

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
          Text(
            label!.toUpperCase(),
            style: const TextStyle(
              fontWeight: FontWeight.bold,
              fontSize: 12,
              letterSpacing: 1.2,
            ),
          ),
          const SizedBox(height: 4),
        ],
        Container(
          height: height,
          padding: padding ?? const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: const Color(0xFFF4F1EA),
            border: Border.all(color: Colors.black, width: 2),
            boxShadow: const [
              BoxShadow(
                color: Colors.black,
                offset: Offset(4, 4),
                blurRadius: 0,
              ),
            ],
          ),
          child: child,
        ),
      ],
    );
  }
}
