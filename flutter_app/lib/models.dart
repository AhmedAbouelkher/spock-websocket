import 'package:equatable/equatable.dart';

class SocketEventType {
  SocketEventType._();
  static const String message = 'message';
  static const String newRoom = 'new_room';
}

// MARK: - SocketEvent
class SocketEvent extends Equatable {
  /// Use `SocketEventType` constants for the type.
  final String type;
  final DateTime timestamp;
  final Map<String, dynamic> data;

  const SocketEvent({
    required this.type,
    required this.timestamp,
    required this.data,
  });

  factory SocketEvent.fromJson(Map<String, dynamic> json) => SocketEvent(
        type: json["type"],
        timestamp: DateTime.parse(json["timestamp"]),
        data: json["data"],
      );

  Map<String, dynamic> toJson() => {
        "type": type,
        "timestamp": timestamp.toIso8601String(),
        "data": data,
      };

  @override
  List<Object> get props => [type, timestamp, data];
}

// MARK: - UserRoomsResponse
class UserRoomsResponse extends Equatable {
  final List<ChatRoom> data;
  final int total;
  final int page;
  final int perPage;
  final int? prev;
  final int? next;
  final int pagesCount;
  final int limit;

  const UserRoomsResponse({
    required this.data,
    required this.total,
    required this.page,
    required this.perPage,
    required this.prev,
    required this.next,
    required this.pagesCount,
    required this.limit,
  });

  factory UserRoomsResponse.fromJson(Map<String, dynamic> json) =>
      UserRoomsResponse(
        data: List<ChatRoom>.from(
          (json["data"] as List).map((x) => ChatRoom.fromJson(x)),
        ),
        total: json["total"],
        page: json["page"],
        perPage: json["per_page"],
        prev: json["prev"],
        next: json["next"],
        pagesCount: json["pages_count"],
        limit: json["limit"],
      );

  Map<String, dynamic> toJson() => {
        "data": List<dynamic>.from(data.map((x) => x.toJson())),
        "total": total,
        "page": page,
        "per_page": perPage,
        "prev": prev,
        "next": next,
        "pages_count": pagesCount,
        "limit": limit,
      };

  @override
  List<Object?> get props {
    return [
      data,
      total,
      page,
      perPage,
      prev,
      next,
      pagesCount,
      limit,
    ];
  }
}

// MARK: - UsersResponse
class UsersResponse extends Equatable {
  final List<AppUser> data;
  final int total;
  final int page;
  final int perPage;
  final int? prev;
  final int? next;
  final int pagesCount;
  final int limit;

  const UsersResponse({
    required this.data,
    required this.total,
    required this.page,
    required this.perPage,
    required this.prev,
    required this.next,
    required this.pagesCount,
    required this.limit,
  });

  factory UsersResponse.fromJson(Map<String, dynamic> json) => UsersResponse(
        data: List<AppUser>.from(
          (json["data"] as List).map((x) => AppUser.fromJson(x)),
        ),
        total: json["total"],
        page: json["page"],
        perPage: json["per_page"],
        prev: json["prev"],
        next: json["next"],
        pagesCount: json["pages_count"],
        limit: json["limit"],
      );

  Map<String, dynamic> toJson() => {
        "data": List<dynamic>.from(data.map((x) => x.toJson())),
        "total": total,
        "page": page,
        "per_page": perPage,
        "prev": prev,
        "next": next,
        "pages_count": pagesCount,
        "limit": limit,
      };

  @override
  List<Object?> get props {
    return [
      data,
      total,
      page,
      perPage,
      prev,
      next,
      pagesCount,
      limit,
    ];
  }
}

class ChatRoomType {
  ChatRoomType._();
  static const String direct = 'private';
  static const String group = 'group';
}

// MARK: - ChatRoom
class ChatRoom extends Equatable {
  final String roomId;
  final String name;

  /// Use `ChatRoomType` constants for the type.
  final String type;
  final DateTime createdAt;
  final int? numberOfUsers;
  final List<AppUser>? users;
  final List<String>? usersIDs;
  final AppUser? otherUser;
  final ChatMessage? lastMessage;

  const ChatRoom({
    required this.roomId,
    required this.name,
    required this.type,
    required this.createdAt,
    required this.numberOfUsers,
    required this.users,
    required this.usersIDs,
    required this.otherUser,
    required this.lastMessage,
  });

  factory ChatRoom.fromJson(Map<String, dynamic> json) => ChatRoom(
        roomId: json["room_id"],
        name: json["name"],
        type: json["type"],
        createdAt: DateTime.parse(json["created_at"]),
        numberOfUsers: json["number_of_users"],
        users: json["users"] == null
            ? null
            : List<AppUser>.from(json["users"].map((x) => AppUser.fromJson(x))),
        usersIDs: json["users_ids"] == null
            ? null
            : List<String>.from(json["users_ids"].map((x) => x)),
        otherUser: json["other_user"] == null
            ? null
            : AppUser.fromJson(json["other_user"]),
        lastMessage: json["last_message"] == null
            ? null
            : ChatMessage.fromJson(json["last_message"]),
      );

  Map<String, dynamic> toJson() => {
        "room_id": roomId,
        "name": name,
        "type": type,
        "created_at": createdAt.toIso8601String(),
        "number_of_users": numberOfUsers,
        "users": users == null
            ? null
            : List<dynamic>.from(users!.map((x) => x.toJson())),
        "users_ids": usersIDs == null ? null : List<dynamic>.from(usersIDs!),
        "other_user": otherUser?.toJson(),
        "last_message": lastMessage?.toJson(),
      };

  @override
  List<Object?> get props {
    return [
      roomId,
      name,
      type,
      createdAt,
      numberOfUsers,
      users,
      usersIDs,
      otherUser,
      lastMessage,
    ];
  }

  ChatRoom copyWith({
    String? roomId,
    String? name,
    String? type,
    DateTime? createdAt,
    int? numberOfUsers,
    List<AppUser>? users,
    List<String>? usersIDs,
    AppUser? otherUser,
    ChatMessage? lastMessage,
  }) {
    return ChatRoom(
      roomId: roomId ?? this.roomId,
      name: name ?? this.name,
      type: type ?? this.type,
      createdAt: createdAt ?? this.createdAt,
      numberOfUsers: numberOfUsers ?? this.numberOfUsers,
      users: users ?? this.users,
      usersIDs: usersIDs ?? this.usersIDs,
      otherUser: otherUser ?? this.otherUser,
      lastMessage: lastMessage ?? this.lastMessage,
    );
  }
}

// MARK: - ChatRoomMessagesResponse
class ChatRoomMessagesResponse extends Equatable {
  final List<ChatMessage> data;
  final int total;
  final int page;
  final int perPage;
  final int? prev;
  final int? next;
  final int pagesCount;
  final int limit;

  const ChatRoomMessagesResponse({
    required this.data,
    required this.total,
    required this.page,
    required this.perPage,
    required this.prev,
    required this.next,
    required this.pagesCount,
    required this.limit,
  });

  factory ChatRoomMessagesResponse.fromJson(Map<String, dynamic> json) =>
      ChatRoomMessagesResponse(
        data: json["data"] == null
            ? []
            : List<ChatMessage>.from(
                (json["data"] as List).map(
                  (x) => ChatMessage.fromJson(x),
                ),
              ),
        total: json["total"],
        page: json["page"],
        perPage: json["per_page"],
        prev: json["prev"],
        next: json["next"],
        pagesCount: json["pages_count"],
        limit: json["limit"],
      );

  Map<String, dynamic> toJson() => {
        "data": List<dynamic>.from(data.map((x) => x.toJson())),
        "total": total,
        "page": page,
        "per_page": perPage,
        "prev": prev,
        "next": next,
        "pages_count": pagesCount,
        "limit": limit,
      };

  @override
  List<Object?> get props {
    return [
      data,
      total,
      page,
      perPage,
      prev,
      next,
      pagesCount,
      limit,
    ];
  }
}

// MARK: - ChatMessage
class ChatMessage extends Equatable {
  final int id;
  final String content;
  final String type;
  final DateTime sentAt;
  final DateTime? editedAt;
  final bool myMessage;
  final String? senderId;
  final String? roomId;
  final AppUser? sentBy;

  const ChatMessage({
    required this.id,
    required this.content,
    required this.type,
    required this.sentAt,
    required this.editedAt,
    required this.myMessage,
    required this.senderId,
    required this.roomId,
    required this.sentBy,
  });

  factory ChatMessage.fromJson(Map<String, dynamic> json) => ChatMessage(
        id: json["id"],
        content: json["content"],
        type: json["type"],
        sentAt: DateTime.parse(json["sent_at"]),
        editedAt: json["edited_at"] == null
            ? null
            : DateTime.parse(json["edited_at"]),
        myMessage: json["my_message"],
        senderId: json["sender_id"],
        roomId: json["room_id"],
        sentBy:
            json["sent_by"] == null ? null : AppUser.fromJson(json["sent_by"]),
      );

  Map<String, dynamic> toJson() => {
        "id": id,
        "content": content,
        "type": type,
        "sent_at": sentAt.toIso8601String(),
        "edited_at": editedAt?.toIso8601String(),
        "my_message": myMessage,
        "sender_id": senderId,
        "room_id": roomId,
        "sent_by": sentBy?.toJson(),
      };

  @override
  List<Object?> get props {
    return [
      id,
      content,
      type,
      sentAt,
      editedAt,
      myMessage,
      senderId,
      roomId,
      sentBy,
    ];
  }
}

// MARK: - AppUser
class AppUser extends Equatable {
  final String id;
  final String name;
  final String profileImageIcon;
  final String email;
  final DateTime createdAt;
  final DateTime updatedAt;

  const AppUser({
    required this.id,
    required this.name,
    required this.profileImageIcon,
    required this.email,
    required this.createdAt,
    required this.updatedAt,
  });

  static AppUser get empty => AppUser(
        id: '',
        name: '',
        profileImageIcon:
            'https://api.dicebear.com/9.x/pixel-art/svg?seed=startrek',
        email: '',
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
      );

  bool get isEmpty => id.isEmpty;

  factory AppUser.fromJson(Map<String, dynamic> json) {
    final id = json["id"] as String;
    return AppUser(
      id: id,
      name: json["name"],
      profileImageIcon: (json["profile_image_icon"] as String?) ??
          "https://api.dicebear.com/9.x/pixel-art/svg?seed=$id",
      email: json["email"],
      createdAt: DateTime.parse(json["created_at"]),
      updatedAt: DateTime.parse(json["updated_at"]),
    );
  }

  Map<String, dynamic> toJson() => {
        "id": id,
        "name": name,
        "profile_image_icon": profileImageIcon,
        "email": email,
        "created_at": createdAt.toIso8601String(),
        "updated_at": updatedAt.toIso8601String(),
      };

  @override
  List<Object?> get props {
    return [
      id,
      name,
      profileImageIcon,
      email,
      createdAt,
      updatedAt,
    ];
  }
}
