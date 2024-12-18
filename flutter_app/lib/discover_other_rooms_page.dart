import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_app/models.dart';
import 'package:flutter_app/utils.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;

class DiscoverOtherChatRoomsPage extends StatefulWidget {
  const DiscoverOtherChatRoomsPage({super.key});

  @override
  State<DiscoverOtherChatRoomsPage> createState() =>
      _DiscoverOtherChatRoomsPageState();
}

class _DiscoverOtherChatRoomsPageState
    extends State<DiscoverOtherChatRoomsPage> {
  bool isLoading = true;
  List<ChatRoom> chatRooms = [];

  @override
  void initState() {
    super.initState();
    loadChatRooms();
  }

  Future<void> loadChatRooms() async {
    isLoading = true;
    if (mounted) setState(() {});

    try {
      final prefs = await SharedPreferences.getInstance();
      var uri = Uri.parse('$API_URL/chat/discover-rooms');
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
      chatRooms = rooms.data;
    } catch (e, t) {
      showError(e, t);
    } finally {
      isLoading = false;
      if (mounted) setState(() {});
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Discover New Chat Rooms'),
      ),
      body: SafeArea(
        child: Builder(builder: (context) {
          if (isLoading) {
            return const Center(child: CircularProgressIndicator());
          }
          if (chatRooms.isEmpty) {
            return const Center(
              child: Text('No chat rooms found, please try again later'),
            );
          }

          return RefreshIndicator(
            onRefresh: loadChatRooms,
            child: ListView.separated(
              padding: const EdgeInsets.fromLTRB(0, 20, 0, 80),
              itemCount: chatRooms.length,
              separatorBuilder: (context, index) => const Divider(height: 8),
              itemBuilder: (context, index) {
                final room = chatRooms[index];
                return ListTile(
                  leading: Container(
                    width: 50,
                    height: 50,
                    clipBehavior: Clip.antiAlias,
                    decoration: BoxDecoration(
                      color: Colors.grey.shade200,
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(Icons.group),
                  ),
                  title: Text(
                    room.name,
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.bold,
                        ),
                  ),
                  onTap: () {
                    showMessage("User ${room.name} tapped");
                  },
                );
              },
            ),
          );
        }),
      ),
    );
  }
}
