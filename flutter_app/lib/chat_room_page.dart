import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_svg/svg.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

import 'models.dart';
import 'utils.dart';

class ChatRoomPage extends StatefulWidget {
  /// The chat room to display messages for, will be null if the chat room is not created yet.
  final ChatRoom? room;

  /// The other user, usually a new user that the current user is chatting with for the first time.
  final AppUser? otherUser;

  const ChatRoomPage({
    super.key,
    this.room,
    this.otherUser,
  }) : assert(room != null || otherUser != null,
            'room or otherUser must be provided');

  @override
  State<ChatRoomPage> createState() => ChatRoomPageState();
}

class ChatRoomPageState extends State<ChatRoomPage> {
  final messageController = TextEditingController();

  bool isLoading = false;
  List<ChatMessage> roomMessages = [];

  StreamSubscription<SocketEvent>? _subscription;

  ChatRoom? get room => widget.room;
  AppUser? get otherUser => widget.otherUser;

  @override
  void initState() {
    super.initState();
    isLoading = room !=
        null; // no need to load messages if the room is not created yet.
    setupSocketListener();
    loadRoomMessages();
  }

  @override
  void dispose() {
    messageController.dispose();
    _subscription?.cancel();
    super.dispose();
  }

  Future<void> loadRoomMessages() async {
    // the users do not have a room yet
    if (room == null) {
      return;
    }

    isLoading = true;
    if (mounted) setState(() {});

    try {
      final prefs = await SharedPreferences.getInstance();
      var uri = Uri.parse('$API_URL/chat/room-messages/${room!.roomId}');
      uri = uri.replace(queryParameters: {
        'page': '1',
        'limit': '100' // we are avoiding pagination
      });
      final response = await http.get(
        uri,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ${prefs.getString('token')}',
        },
      );
      if (response.statusCode != HttpStatus.ok) {
        throw HttpException(response);
      }
      final responseBody = jsonDecode(response.body);
      final rooms = ChatRoomMessagesResponse.fromJson(responseBody);
      roomMessages = rooms.data;
    } catch (e, t) {
      showError(e, t);
    } finally {
      isLoading = false;
      if (mounted) setState(() {});
    }
  }

  Future<void> setupSocketListener() async {
    final socket = socketConnection;
    if (socket == null) {
      showError('Socket connection is not available');
      return;
    }
    _subscription = socket.events!.listen((event) {
      if (event.type != SocketEventType.message) return;
      final message = ChatMessage.fromJson(event.data);
      roomMessages = [message, ...roomMessages];
      if (mounted) {
        setState(() {});
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: () => FocusScope.of(context).unfocus(),
      child: Scaffold(
        appBar: AppBar(
          title: Text((room?.name ?? otherUser?.name) ?? 'Chat Room'),
          actions: [
            IconButton(
              icon: const Icon(Icons.refresh),
              tooltip: 'Refresh messages',
              onPressed: loadRoomMessages,
            ),
          ],
        ),
        body: SafeArea(
          child: Builder(builder: (context) {
            if (isLoading) {
              return const Center(child: CircularProgressIndicator());
            }

            return Column(
              children: [
                Expanded(
                  child: Builder(builder: (context) {
                    if (roomMessages.isEmpty) {
                      if (otherUser != null) {
                        return Center(
                          child: Text(
                            'No messages found, start chatting with ${otherUser!.name}.',
                          ),
                        );
                      }
                      return const Center(
                        child: Text('Start messaging with the other users'),
                      );
                    }

                    return ListView.builder(
                      padding: const EdgeInsets.fromLTRB(15, 80, 15, 10),
                      reverse: true,
                      itemCount: roomMessages.length,
                      itemBuilder: (context, index) {
                        final message = roomMessages[index];
                        bool hasTheSameSenderAsPrevious = false;
                        bool hasTheSameSenderAsNext = false;
                        if (index > 0) {
                          hasTheSameSenderAsPrevious =
                              roomMessages[index - 1].sentBy?.id ==
                                  message.sentBy?.id;
                        }
                        if (index < roomMessages.length - 1) {
                          hasTheSameSenderAsNext =
                              roomMessages[index + 1].sentBy?.id ==
                                  message.sentBy?.id;
                        }
                        return _ChatMessageListTile(
                          message: message,
                          hasTheSameSenderAsPrevious:
                              hasTheSameSenderAsPrevious,
                          hasTheSameSenderAsNext: hasTheSameSenderAsNext,
                        );
                      },
                    );
                  }),
                ),
                const Divider(height: 1, color: Colors.black),
                Container(
                  padding: const EdgeInsets.all(10),
                  child: Row(
                    children: [
                      const SizedBox(width: 10),
                      Expanded(
                        child: TextField(
                          controller: messageController,
                          autocorrect: false,
                          enableSuggestions: false,
                          decoration: InputDecoration(
                            isDense: true,
                            contentPadding:
                                EdgeInsetsDirectional.fromSTEB(15, 15, 0, 15),
                            hintText: room == null
                                ? 'Start messaging with ${otherUser?.name}'
                                : 'Type a message',
                            focusedBorder: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(30),
                              borderSide: BorderSide(
                                color: Colors.blue,
                                width: 2.0,
                              ),
                            ),
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(30),
                              borderSide: BorderSide(
                                color: Colors.grey,
                                width: 1,
                              ),
                            ),
                          ),
                          onSubmitted: (value) {
                            final message = value.trim();
                            if (message.isEmpty) return;
                            messageController.clear();
                            socketConnection?.send(
                              roomId: room?.roomId,
                              otherUserId: otherUser?.id,
                              message: message,
                            );
                          },
                        ),
                      ),
                      const SizedBox(width: 10),
                      IconButton(
                        icon: const Icon(Icons.send),
                        onPressed: () {
                          final message = messageController.text.trim();
                          if (message.isEmpty) return;
                          messageController.clear();
                          socketConnection?.send(
                            roomId: room?.roomId,
                            otherUserId: otherUser?.id,
                            message: message,
                          );
                        },
                      ),
                    ],
                  ),
                ),
              ],
            );
          }),
        ),
      ),
    );
  }
}

class _ChatMessageListTile extends StatelessWidget {
  final ChatMessage message;
  final bool hasTheSameSenderAsPrevious;
  final bool hasTheSameSenderAsNext;
  const _ChatMessageListTile({
    required this.message,
    this.hasTheSameSenderAsPrevious = false,
    this.hasTheSameSenderAsNext = false,
  });

  static const _kRadius = Radius.circular(8);

  @override
  Widget build(BuildContext context) {
    final sentBy = message.sentBy;
    final isSentByMe = message.myMessage;
    if (isSentByMe) {
      return Align(
        alignment: AlignmentDirectional.centerEnd,
        child: Container(
          padding: const EdgeInsets.all(10),
          margin: hasTheSameSenderAsNext
              ? const EdgeInsets.only(top: 5)
              : const EdgeInsets.only(top: 10),
          decoration: BoxDecoration(
            color: Colors.blue.shade100,
            borderRadius: _createOtherUserMessageBorderRadius(),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                message.content,
                textAlign: TextAlign.start,
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
              ),
              const SizedBox(height: 8),
              Text(dateToTimeFormatter(context, message.sentAt)),
            ],
          ),
        ),
      );
    }
    return Padding(
      padding: hasTheSameSenderAsNext
          ? EdgeInsets.zero
          : const EdgeInsets.only(top: 10),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            clipBehavior: Clip.antiAlias,
            decoration: BoxDecoration(
              color: Colors.grey.shade200,
              shape: BoxShape.circle,
            ),
            child: SvgPicture.network(sentBy?.profileImageIcon ?? ""),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (!hasTheSameSenderAsNext) ...[
                  Text(
                    sentBy?.name ?? 'App User',
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ],
                const SizedBox(height: 5),
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: Colors.grey.shade200,
                    borderRadius: _createOtherUserMessageBorderRadius(),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        message.content,
                        textAlign: TextAlign.start,
                        style: TextStyle(
                            fontSize: 16, fontWeight: FontWeight.w600),
                      ),
                      const SizedBox(height: 8),
                      Text(dateToTimeFormatter(context, message.sentAt)),
                    ],
                  ),
                )
              ],
            ),
          ),
        ],
      ),
    );
  }

  BorderRadiusGeometry _createOtherUserMessageBorderRadius() {
    Radius topStart = _kRadius;
    Radius topEnd = _kRadius;
    Radius bottomStart = _kRadius;
    Radius bottomEnd = _kRadius;

    if (hasTheSameSenderAsNext) {
      topStart = Radius.zero;
      topEnd = Radius.zero;
    }
    if (hasTheSameSenderAsPrevious) {
      bottomStart = Radius.zero;
      bottomEnd = Radius.zero;
    }

    return BorderRadiusDirectional.only(
      topStart: topStart,
      topEnd: topEnd,
      bottomStart: bottomStart,
      bottomEnd: bottomEnd,
    );
  }
}
