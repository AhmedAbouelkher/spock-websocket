import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_app/models.dart';
import 'package:flutter_app/utils.dart';
import 'package:flutter_svg/svg.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;

import 'chat_room_page.dart';

class DiscoverOtherUsersPage extends StatefulWidget {
  const DiscoverOtherUsersPage({super.key});

  @override
  State<DiscoverOtherUsersPage> createState() => _DiscoverOtherUsersPageState();
}

class _DiscoverOtherUsersPageState extends State<DiscoverOtherUsersPage> {
  bool isLoading = true;
  List<AppUser> otherUsers = [];

  @override
  void initState() {
    super.initState();
    loadOtherUsers();
  }

  Future<void> loadOtherUsers() async {
    isLoading = true;
    if (mounted) setState(() {});

    try {
      final prefs = await SharedPreferences.getInstance();
      var uri = Uri.parse('$API_URL/chat/discover-users');
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
      final users = UsersResponse.fromJson(responseBody);
      otherUsers = users.data;
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
        title: const Text('Discover Users'),
      ),
      body: SafeArea(
        child: Builder(builder: (context) {
          if (isLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (otherUsers.isEmpty) {
            return const Center(
              child: Text('No users found'),
            );
          }

          return RefreshIndicator(
            onRefresh: loadOtherUsers,
            child: ListView.separated(
              padding: const EdgeInsets.fromLTRB(0, 20, 0, 80),
              itemCount: otherUsers.length,
              separatorBuilder: (context, index) => const Divider(height: 8),
              itemBuilder: (context, index) {
                final user = otherUsers[index];
                return ListTile(
                  leading: Container(
                    width: 60,
                    height: 60,
                    clipBehavior: Clip.antiAlias,
                    decoration: BoxDecoration(
                      color: Colors.grey.shade200,
                      shape: BoxShape.circle,
                    ),
                    child: SvgPicture.network(user.profileImageIcon),
                  ),
                  title: Text(
                    user.name,
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.bold,
                        ),
                  ),
                  trailing: const Icon(Icons.chat_outlined),
                  onTap: () {
                    Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => ChatRoomPage(otherUser: user),
                      ),
                    );
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
