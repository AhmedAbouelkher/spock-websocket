// ignore_for_file: constant_identifier_names

import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_app/models.dart';
import 'package:fluttertoast/fluttertoast.dart';
import 'package:http/http.dart' as http;

import 'websocket.dart';

const API_URL = 'http://localhost:4444/api/v1';
const WS_URL = 'ws://localhost:4444/api/v1/ws/chat';

AppUser currentUser = AppUser.empty;
// Do not expose a socket connection to the global scope in production, use a service instead
EnhancedWebSocket? socketConnection;

class HttpException implements Exception {
  final http.Response response;
  const HttpException(this.response);
  @override
  String toString() => _getErrorMessage();
  String _getErrorMessage() {
    final body = response.body;
    final json = jsonDecode(body);
    String errorMessage = 'Unknown network error';
    if (json is Map) {
      errorMessage = json['message'] ?? json['error'] ?? response.reasonPhrase;
    }
    return errorMessage;
  }
}

void showError(Object err, [StackTrace? stackTrace]) {
  String errMsg = err.toString();
  if (err is PlatformException) {
    errMsg = "${err.code}: ${err.message}";
  }
  // for debugging
  debugPrint("ERROR MESSAGE: $errMsg");
  debugPrintStack(stackTrace: stackTrace ?? StackTrace.current);
// display error message to the user
  Fluttertoast.showToast(
    msg: errMsg,
    toastLength: Toast.LENGTH_LONG,
    textColor: Colors.white,
    fontSize: 16,
    backgroundColor: Colors.red.shade400,
  );
}

Future<void> showMessage(String msg, {Toast? toastLength}) async {
  await Fluttertoast.showToast(
    msg: msg,
    toastLength: toastLength ?? Toast.LENGTH_SHORT,
    textColor: Colors.white,
    backgroundColor: Colors.purple,
  );
}

String dateToTimeFormatter(BuildContext context, DateTime date) {
  return TimeOfDay.fromDateTime(date.toLocal()).format(context);
}

String formateDate(DateTime date) {
  return '${date.day}/${date.month}/${date.year}';
}
