import 'dart:math' as math;

import 'package:flutter/material.dart';

/// 环形指标图：圆环进度 + 中心主数值，下方标签与子文本。
/// 用于看板总览的负载/内存/CPU/磁盘四联环（参考 HexHub 移动端布局）。
class MetricGauge extends StatelessWidget {
  final String label;
  final String centerText;
  final String subText;

  /// 0~1 的环进度
  final double fraction;
  final Color color;

  const MetricGauge({
    super.key,
    required this.label,
    required this.centerText,
    required this.subText,
    required this.fraction,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        SizedBox(
          width: 76,
          height: 76,
          child: CustomPaint(
            painter: _GaugePainter(
              fraction: fraction.clamp(0, 1),
              color: color,
              trackColor:
                  Theme.of(context).colorScheme.onSurface.withValues(alpha: 0.12),
            ),
            child: Center(
              child: Text(
                centerText,
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                      fontSize: 15,
                    ),
                textAlign: TextAlign.center,
              ),
            ),
          ),
        ),
        const SizedBox(height: 6),
        Text(label, style: Theme.of(context).textTheme.bodyMedium),
        const SizedBox(height: 2),
        Text(
          subText,
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Theme.of(context)
                    .colorScheme
                    .onSurface
                    .withValues(alpha: 0.6),
                fontSize: 11,
              ),
          textAlign: TextAlign.center,
          // 每环列宽只有约 85px，一行放不下完整用量，允许两行避免被截断
          maxLines: 2,
          overflow: TextOverflow.ellipsis,
        ),
      ],
    );
  }
}

class _GaugePainter extends CustomPainter {
  final double fraction;
  final Color color;
  final Color trackColor;

  _GaugePainter({
    required this.fraction,
    required this.color,
    required this.trackColor,
  });

  @override
  void paint(Canvas canvas, Size size) {
    const strokeWidth = 7.0;
    final center = Offset(size.width / 2, size.height / 2);
    final radius = (size.width - strokeWidth) / 2;

    final track = Paint()
      ..color = trackColor
      ..style = PaintingStyle.stroke
      ..strokeWidth = strokeWidth
      ..strokeCap = StrokeCap.round;
    canvas.drawCircle(center, radius, track);

    if (fraction > 0) {
      final arc = Paint()
        ..color = color
        ..style = PaintingStyle.stroke
        ..strokeWidth = strokeWidth
        ..strokeCap = StrokeCap.round;
      canvas.drawArc(
        Rect.fromCircle(center: center, radius: radius),
        -math.pi / 2,
        2 * math.pi * fraction,
        false,
        arc,
      );
    }
  }

  @override
  bool shouldRepaint(_GaugePainter oldDelegate) {
    return oldDelegate.fraction != fraction ||
        oldDelegate.color != color ||
        oldDelegate.trackColor != trackColor;
  }
}
