import 'package:flutter/material.dart';

/// Wait for current capabilities before constructing the popup route. Updating
/// the screen after opening a PopupMenuButton would leave its items stale.
class CheckedPopupMenuButton<T> extends StatefulWidget {
  const CheckedPopupMenuButton({
    super.key,
    required this.check,
    required this.enabled,
    required this.itemBuilder,
    required this.onSelected,
    required this.child,
    this.initialValue,
    this.tooltip,
    this.constraints,
  });

  final Future<bool> Function() check;
  final bool enabled;
  final PopupMenuItemBuilder<T> itemBuilder;
  final PopupMenuItemSelected<T> onSelected;
  final Widget child;
  final T? initialValue;
  final String? tooltip;
  final BoxConstraints? constraints;

  @override
  State<CheckedPopupMenuButton<T>> createState() =>
      _CheckedPopupMenuButtonState<T>();
}

class _CheckedPopupMenuButtonState<T> extends State<CheckedPopupMenuButton<T>> {
  final _popup = GlobalKey<PopupMenuButtonState<T>>();
  bool _checking = false;

  Future<void> _open() async {
    if (_checking || !widget.enabled) return;
    setState(() => _checking = true);
    try {
      final ready = await widget.check();
      if (!mounted || !widget.enabled || !ready) return;
      // Let model notifications rebuild itemBuilder/initialValue before opening.
      await WidgetsBinding.instance.endOfFrame;
      if (mounted && widget.enabled) _popup.currentState?.showButtonMenu();
    } finally {
      if (mounted) setState(() => _checking = false);
    }
  }

  @override
  Widget build(BuildContext context) => PopupMenuButton<T>(
    key: _popup,
    enabled: false,
    tooltip: widget.tooltip,
    initialValue: widget.initialValue,
    onSelected: widget.onSelected,
    padding: EdgeInsets.zero,
    menuPadding: const EdgeInsets.symmetric(vertical: 4),
    constraints: widget.constraints,
    itemBuilder: widget.itemBuilder,
    child: InkWell(
      onTap: widget.enabled && !_checking ? _open : null,
      child: widget.child,
    ),
  );
}
