import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:web_socket_channel/io.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

import 'package:flutter_app/models.dart';

const Duration _kPingInterval = Duration(seconds: 10);
const Duration _kConnectionTimeout = Duration(seconds: 5);

class CloseCode {
  CloseCode._();
  static const int normalClosure = 1000;
  static const int goingAway = 1001;
  static const int protocolError = 1002;
  static const int unsupportedData = 1003;
  static const int noStatusReceived = 1005;
  static const int abnormalClosure = 1006;
  static const int invalidFramePayloadData = 1007;
  static const int policyViolation = 1008;
  static const int messageTooBig = 1009;
  static const int mandatoryExtension = 1010;
  static const int internalServerError = 1011;
  static const int serviceRestart = 1012;
  static const int tryAgainLater = 1013;
  static const int tlsHandshake = 1015;
}

enum SocketConnectionState {
  opening,
  open,
  closing,
  closed,
  reconnecting,
}

class EnhancedWebSocket {
  final String url;
  final Map<String, String> queryParameters;

  EnhancedWebSocket(
    this.url, {
    this.queryParameters = const {},
  });

  WebSocketChannel? _channel;
  StreamController<SocketEvent>? _eventsController;
  Stream<SocketEvent>? get events => _eventsController?.stream;

  bool _clientClosedConn = false;
  SocketConnectionState _connectionState = SocketConnectionState.closed;
  SocketConnectionState get connectionState => _connectionState;

  bool get _isOpen => _connectionState == SocketConnectionState.open;
  bool get _isClosing => _connectionState == SocketConnectionState.closing;
  bool get _isClosed => _connectionState == SocketConnectionState.closed;
  bool get _isReconnecting =>
      _connectionState == SocketConnectionState.reconnecting;

  Future<void> reconnect() async {
    if (_isClosed) {
      await connect(force: true);
      return;
    }
    _connectionState = SocketConnectionState.closing;
    await _channel?.sink.close(CloseCode.normalClosure, 'reconnecting');
    _connectionState = SocketConnectionState.closed;
    await connect(force: true);
  }

  Future<void> connect({bool force = false}) async {
    if ((_isOpen || _clientClosedConn) && !force) return;
    if (force) {
      _clientClosedConn = false;
    }
    await _connect();
  }

  Future<void> _connect() async {
    void attemptToReconnect() {
      if (_clientClosedConn || _isReconnecting || _isClosing) return;
      debugPrint(
        'WS Attempting to reconnect due to '
        'CloseCode: ${_channel?.closeCode} and CloseReason: ${_channel?.closeReason}',
      );
      _channel = null;
      _connectionState = SocketConnectionState.reconnecting;
      Future.delayed(_kConnectionTimeout, reconnect);
    }

    try {
      final uri = Uri.parse(url).replace(queryParameters: queryParameters);
      _channel = IOWebSocketChannel.connect(
        uri.toString(),
        pingInterval: _kPingInterval,
        connectTimeout: _kConnectionTimeout,
      );
      debugPrint('[EnhancedWebSocket] WS Connecting to: $url');
      await _channel!.ready;

      debugPrint('[EnhancedWebSocket] WS Connected');
      _connectionState = SocketConnectionState.open;

      _eventsController ??= StreamController<SocketEvent>.broadcast();
      _channel!.stream.listen(
        (dynamic event) {
          switch (event.runtimeType) {
            case const (String):
              final data = jsonDecode(event as String);
              final socketEvent = SocketEvent.fromJson(data);
              debugPrint(
                  '[EnhancedWebSocket] New Event Received of type ${socketEvent.type}');
              _eventsController?.add(socketEvent);
              break;
            case const (Uint8List):
              debugPrint(
                  '[EnhancedWebSocket] New Event Received of type Uint8List');
              break;
          }
        },
        onDone: attemptToReconnect,
        cancelOnError: true,
      );
    } catch (_) {
      attemptToReconnect();
    }
  }

  Future<void> close([int? code, String? reason]) async {
    if (_clientClosedConn) return;
    _clientClosedConn = true;
    _connectionState = SocketConnectionState.closing;
    _channel?.sink.close(code, reason).whenComplete(() {
      debugPrint('[EnhancedWebSocket] WS Closed');

      _channel = null;
      _connectionState = SocketConnectionState.closed;
      _eventsController?.close();
      _eventsController = null;
    });
  }

  void send({
    String? roomId,
    String? otherUserId,
    required String message,
  }) async {
    if (_isClosed) return;
    assert(_channel != null, 'Connect to the socket first');
    assert(roomId != null || otherUserId != null,
        'roomId or otherUserId must be provided');

    final data = jsonEncode({
      'event': 'message',
      'data': {
        if (roomId != null) 'room_id': roomId,
        if (otherUserId != null) 'other_user_id': otherUserId,
        'content': message,
      },
    });
    _channel!.sink.add(data);
  }
}
