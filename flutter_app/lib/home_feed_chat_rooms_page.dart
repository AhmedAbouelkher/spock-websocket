import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_svg/svg.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

import 'package:flutter_app/websocket.dart';

import 'chat_room_page.dart';
import 'discover_other_rooms_page.dart';
import 'discover_other_users_page.dart';
import 'login_register_profile_pages.dart';
import 'models.dart';
import 'utils.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  StreamSubscription<SocketEvent>? _subscription;
  bool isLoading = true;
  List<ChatRoom> userRooms = [];

  bool checkboxesMode = false;
  Set<ChatRoom> selectedRooms = {};

  @override
  void initState() {
    super.initState();
    initWebsocket();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      loadChatRooms();
    });
  }

  @override
  void dispose() {
    _subscription?.cancel();
    socketConnection?.close(); // do not do this in production, use a service
    super.dispose();
  }

  Future<void> loadChatRooms() async {
    isLoading = true;
    if (mounted) setState(() {});

    try {
      final prefs = await SharedPreferences.getInstance();
      var uri = Uri.parse('$API_URL/chat/rooms');
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
      final rooms = UserRoomsResponse.fromJson(responseBody);
      userRooms = rooms.data;
    } catch (e, t) {
      showError(e, t);
    } finally {
      isLoading = false;
      if (mounted) setState(() {});
    }
  }

  Future<void> createGroupFromSelection() async {
    final result = await showDialog<String>(
      context: context,
      builder: (context) {
        String groupName = '';
        return AlertDialog(
          title: const Text('Create Group'),
          content: TextField(
            onChanged: (value) => groupName = value,
            decoration: const InputDecoration(
              hintText: 'Group Name',
              labelText: 'Enter group name',
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text('Cancel'),
            ),
            TextButton(
              onPressed: () => Navigator.of(context).pop(groupName),
              child: const Text('Create'),
            ),
          ],
        );
      },
    );
    if (result != null && result.isNotEmpty) {
      final usersIds = selectedRooms
          .map((r) => r.otherUser?.id)
          .where((id) => id != null)
          .toList();
      try {
        final prefs = await SharedPreferences.getInstance();
        final response = await http.post(
          Uri.parse('$API_URL/chat/create-group-room'),
          headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ${prefs.getString('token')}',
          },
          body: jsonEncode({
            'name': result,
            'other_users_ids': usersIds,
          }),
        );
        if (response.statusCode != HttpStatus.ok) {
          throw HttpException(response);
        }
        final responseBody = jsonDecode(response.body);
        final room = ChatRoom.fromJson(responseBody);
        showMessage('Group created successfully');
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (_) => ChatRoomPage(room: room),
          ),
        );
        userRooms.insert(0, room);
      } catch (e, t) {
        showError(e, t);
      } finally {
        checkboxesMode = false;
        selectedRooms.clear();
        if (mounted) setState(() {});
      }
    }
  }

  // do not do this in production, use a service
  Future<void> initWebsocket() async {
    final prefs = await SharedPreferences.getInstance();
    final token = prefs.getString('token');
    assert(token != null, 'Token is required');
    socketConnection = EnhancedWebSocket(
      WS_URL,
      queryParameters: {"token": token!},
    );
    await socketConnection?.connect();
    _subscription = socketConnection?.events?.listen(socketListener);
  }

  Future<void> reconnectWebsocket() async {
    await socketConnection?.reconnect();

    _subscription?.cancel();
    _subscription = socketConnection?.events?.listen(socketListener);
  }

  void socketListener(SocketEvent event) {
    switch (event.type) {
      case SocketEventType.message:
        final message = ChatMessage.fromJson(event.data);
        final roomId = message.roomId;
        // pump this room to the top
        final roomIndex = userRooms.indexWhere((r) => r.roomId == roomId);
        final isRoomFound = roomIndex != -1;
        if (isRoomFound) {
          var room = userRooms.removeAt(roomIndex);
          room = room.copyWith(lastMessage: message);
          userRooms.insert(0, room);
        } else {
          debugPrint('Received message for unknown room: $roomId');
        }
        break;

      case SocketEventType.newRoom:
        final room = ChatRoom.fromJson(event.data);
        userRooms.insert(0, room);
        break;
    }

    if (mounted) {
      setState(() {});
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      drawer: ProfileDrawer(),
      appBar: AppBar(
        title: const Text('Home Feed'),
        actions: [
          IconButton(
            onPressed: reconnectWebsocket,
            icon: const Icon(Icons.refresh),
          ),
        ],
      ),
      body: SafeArea(
        child: RefreshIndicator(
          onRefresh: loadChatRooms,
          child: CustomScrollView(
            slivers: [
              SliverPadding(
                padding: const EdgeInsets.all(20),
                sliver: SliverToBoxAdapter(
                  child: Row(
                    children: [
                      Expanded(
                        child: _HomeBanner(
                          title: "Discover New Chat Rooms",
                          icon: const Icon(Icons.groups_2_outlined),
                          onTap: () {
                            Navigator.of(context).push(
                              MaterialPageRoute(
                                builder: (_) =>
                                    const DiscoverOtherChatRoomsPage(),
                              ),
                            );
                          },
                        ),
                      ),
                      const SizedBox(width: 10),
                      Expanded(
                        child: _HomeBanner(
                          title: "Discover Other Users",
                          icon: const Icon(Icons.person),
                          onTap: () {
                            Navigator.of(context).push(
                              MaterialPageRoute(
                                builder: (_) => const DiscoverOtherUsersPage(),
                              ),
                            );
                          },
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              if (checkboxesMode) ...[
                SliverPadding(
                  padding: const EdgeInsets.only(bottom: 10),
                  sliver: SliverToBoxAdapter(
                    child: Row(
                      children: [
                        const SizedBox(width: 15),
                        IconButton(
                          onPressed: () {
                            checkboxesMode = false;
                            selectedRooms.clear();
                            if (mounted) {
                              setState(() {});
                            }
                          },
                          icon: const Icon(
                            Icons.indeterminate_check_box_outlined,
                          ),
                        ),
                        const Spacer(),
                        if (selectedRooms.isNotEmpty) ...[
                          TextButton(
                            onPressed: createGroupFromSelection,
                            child: Text('Create Group'),
                          ),
                        ]
                      ],
                    ),
                  ),
                ),
              ],
              Builder(
                builder: (context) {
                  if (isLoading) {
                    return const SliverToBoxAdapter(
                      child: Center(child: CircularProgressIndicator()),
                    );
                  }
                  if (userRooms.isEmpty) {
                    return const SliverToBoxAdapter(
                      child: Center(
                        child: Text(
                          'You did not start a chat or join any room\n'
                          'Start a chat or join a room to see them here',
                        ),
                      ),
                    );
                  }

                  return _buildRoomsSection();
                },
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildRoomsSection() {
    return SliverList.separated(
      separatorBuilder: (context, index) => const Divider(height: 0),
      itemCount: userRooms.length,
      itemBuilder: (context, index) {
        final room = userRooms[index];
        final latestMessage = room.lastMessage;
        Widget? leading;
        if (room.type == 'private') {
          final otherUser = room.otherUser;
          leading = SvgPicture.network(otherUser?.profileImageIcon ?? "");
        } else {
          leading = const Icon(Icons.group);
        }

        final tileLeading = Container(
          width: 50,
          height: 50,
          clipBehavior: Clip.antiAlias,
          decoration: BoxDecoration(
            color: Colors.grey.shade200,
            shape: BoxShape.circle,
          ),
          child: leading,
        );
        final tileTitle = Text(
          room.name,
          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.bold,
              ),
        );
        final tileSubtitle = Builder(
          builder: (context) {
            if (latestMessage == null) {
              return const SizedBox.shrink();
            }
            String content = latestMessage.content;
            if (latestMessage.myMessage) {
              content = 'You: $content';
            } else {
              final name = latestMessage.sentBy?.name;
              content = name != null ? '$name: $content' : content;
            }
            return Text(
              content,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              textAlign: TextAlign.start,
            );
          },
        );

        if (checkboxesMode && room.type != ChatRoomType.group) {
          final isSelected = selectedRooms.contains(room);
          return CheckboxListTile(
            title: tileTitle,
            subtitle: tileSubtitle,
            secondary: tileLeading,
            value: isSelected,
            onChanged: (value) {
              if (value == true) {
                selectedRooms.add(room);
              } else {
                selectedRooms.remove(room);
              }
              if (mounted) {
                setState(() {});
              }
            },
          );
        }

        return InkWell(
          onLongPress: () {
            if (room.type == ChatRoomType.group) return;
            checkboxesMode = true;
            selectedRooms.add(room);
            if (mounted) {
              setState(() {});
            }
          },
          child: ListTile(
            leading: tileLeading,
            title: tileTitle,
            subtitle: tileSubtitle,
            onTap: () {
              Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (_) => ChatRoomPage(room: room),
                ),
              );
            },
          ),
        );
      },
    );
  }
}

class _HomeBanner extends StatelessWidget {
  final String title;
  final Widget icon;
  final VoidCallback? onTap;
  const _HomeBanner({
    required this.title,
    required this.icon,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(10),
      child: Card(
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(10),
          side: BorderSide(color: Colors.grey.shade100),
        ),
        shadowColor: Colors.black45,
        elevation: 2,
        child: Padding(
          padding: const EdgeInsets.all(20.0),
          child: Column(
            children: [
              IconTheme(
                data: const IconThemeData(size: 50),
                child: icon,
              ),
              const SizedBox(height: 10),
              Text(
                title,
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.titleMedium,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
