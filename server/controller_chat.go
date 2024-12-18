package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func DiscoverUsers(c *fiber.Ctx, u *User) (*PaginatedData[User], error) {
	tx := DB().Where("id != ?", u.ID)
	// get users that we do not have any private channel with
	o, err := Paginate(c, User{}, tx, func(tx *gorm.DB) *gorm.DB {
		return tx.Order("RANDOM()")
		// return tx.Order("name ASC")
	})
	if err != nil {
		return nil, err
	}
	return o, nil
}

func DiscoverRooms(c *fiber.Ctx, u *User) (*PaginatedData[ChatRoomResource], error) {
	tx := DB().Where("NOT (? = ANY(users_ids))", u.ID).
		Where("array_length(users_ids, 1) > 2").
		Where("peer_to_peer = FALSE")

	o, err := Paginate(c, ChatRoom{}, tx, func(tx *gorm.DB) *gorm.DB {
		return tx.Order("RANDOM()")
	})
	if err != nil {
		return nil, err
	}

	// get other user in the room (the one who is not the current user)
	usersIds := []string{}
	for _, v := range o.Data {
		for _, id := range v.UsersIDs {
			usersIds = append(usersIds, id)
		}
	}
	roomUsers := []*User{}
	if err := DB().Where("id IN ?", usersIds).Find(&roomUsers).Error; err != nil {
		return nil, err
	}
	roomUsersIndexed := map[string]*User{}
	for _, v := range roomUsers {
		roomUsersIndexed[v.ID.String()] = v
	}

	newO, err := TransformPaginatedData(o, func(data ChatRoom) (ChatRoomResource, error) {
		resource := ChatRoomResource{
			RoomID:        data.ID,
			Type:          RTGroup,
			CreatedAt:     data.CreatedAt,
			Name:          data.Name,
			NumberOfUsers: IntVar(len(data.UsersIDs)),
			UserIDs:       data.UsersIDs,
		}
		for _, id := range data.UsersIDs {
			if id != u.ID.String() {
				resource.Users = append(resource.Users, *roomUsersIndexed[id])
			}
		}
		return resource, nil
	})

	return newO, err
}

func GetRoomsByUserID(c *fiber.Ctx, u *User) (*PaginatedData[ChatRoomResource], error) {
	tx := DB().Where("? = ANY(users_ids)", u.ID)

	o, err := Paginate(c, ChatRoom{}, tx, func(tx *gorm.DB) *gorm.DB {
		return tx.Joins("LatestMessage.CreatedBy").Order(`CASE
		WHEN "LatestMessage"."id" IS NOT NULL THEN "LatestMessage"."created_at"
		ELSE "chat_rooms"."updated_at"
	END DESC`)
	})
	if err != nil {
		return nil, err
	}

	// get other user in the room (the one who is not the current user)
	usersIds := []string{}
	for _, v := range o.Data {
		if *v.PeerToPeer {
			for _, id := range v.UsersIDs {
				if id != u.ID.String() {
					usersIds = append(usersIds, id)
				}
			}
		}
	}
	otherUsers := []*User{}
	if err := DB().Where("id IN ?", usersIds).Find(&otherUsers).Error; err != nil {
		return nil, err
	}
	otherUsersIndexed := map[string]*User{}
	for _, v := range otherUsers {
		otherUsersIndexed[v.ID.String()] = v
	}

	newO, err := TransformPaginatedData(o, func(data ChatRoom) (ChatRoomResource, error) {
		isPrivate := len(data.UsersIDs) == 2 && *data.PeerToPeer
		crType := RTGroup
		numberOfUsers := IntVar(len(data.UsersIDs))
		name := data.Name
		var otherUser *User
		if isPrivate {
			crType = RTPrivate
			numberOfUsers = nil
			for _, id := range data.UsersIDs {
				if id != u.ID.String() {
					otherUser = otherUsersIndexed[id]
					break
				}
			}
			name = otherUser.Name
		}
		var lastMessage *SentMessageResource
		if data.LatestMessage != nil {
			msg := data.LatestMessage
			isMyMessage := msg.CreatedByID == u.ID
			lastMessage = &SentMessageResource{
				ID:        msg.ID,
				Content:   msg.Content,
				Type:      msg.Type,
				SentAt:    msg.CreatedAt,
				EditedAt:  msg.EditedAt,
				MyMassage: isMyMessage,
				SenderID:  &msg.CreatedByID,
			}
			if user := msg.CreatedBy; user != nil {
				lastMessage.User = *user
			}
		}
		return ChatRoomResource{
			RoomID:        data.ID,
			Type:          crType,
			CreatedAt:     data.CreatedAt,
			Name:          name,
			NumberOfUsers: numberOfUsers,
			OtherUser:     otherUser,
			LastMessage:   lastMessage,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return newO, nil
}

type CreateGroupRoomInput struct {
	U             *User
	Name          string   `json:"name"`
	OtherUsersIDs []string `json:"user_ids"`
}

func CreateGroupRoom(in *CreateGroupRoomInput) (any, error) {
	// check if current user is one of the other users ids
	hasCurrentUser := false
	for _, id := range in.OtherUsersIDs {
		if id == in.U.ID.String() {
			hasCurrentUser = true
			break
		}
	}
	if hasCurrentUser {
		return nil, fiber.NewError(fiber.StatusBadRequest, "cannot create room with yourself in it")
	}
	{
		var usersCount int64
		if err := DB().Model(&User{}).
			Where("id IN ?", in.OtherUsersIDs).
			Count(&usersCount).Error; err != nil {
			return nil, err
		}
		if usersCount != int64(len(in.OtherUsersIDs)) {
			return nil, fiber.NewError(fiber.StatusBadRequest, "one or more users not found")
		}
	}
	allRoomUsersIds := append(in.OtherUsersIDs, in.U.ID.String())
	// check if such room already exists
	tx := DB()
	room := &ChatRoom{}
	if err := tx.
		Where("users_ids && ?::VARCHAR(255)[]", StringArray(allRoomUsersIds)).
		Where("array_length(users_ids, 1) = ?", len(allRoomUsersIds)).
		Where("peer_to_peer = FALSE").
		Joins("LatestMessage.CreatedBy").
		First(room).Error; err != nil &&
		!errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	roomAlreadyExists := !room.ID.IsEmpty()
	if roomAlreadyExists {
		var lastMessage *SentMessageResource
		if msg := room.LatestMessage; msg != nil {
			isMyMessage := msg.CreatedByID == in.U.ID
			lastMessage = &SentMessageResource{
				ID:        msg.ID,
				Content:   msg.Content,
				Type:      msg.Type,
				SentAt:    msg.CreatedAt,
				EditedAt:  msg.EditedAt,
				MyMassage: isMyMessage,
				SenderID:  &msg.CreatedByID,
			}
			if user := msg.CreatedBy; user != nil {
				lastMessage.User = *user
			}
		}
		return ChatRoomResource{
			RoomID:        room.ID,
			Name:          room.Name,
			Type:          RTGroup,
			NumberOfUsers: IntVar(len(allRoomUsersIds)),
			UserIDs:       allRoomUsersIds,
			LastMessage:   lastMessage,
			CreatedAt:     room.CreatedAt,
		}, nil
	}
	room = &ChatRoom{
		UsersIDs:   allRoomUsersIds,
		Name:       in.Name,
		UsersLimit: 99,
		PeerToPeer: BoolVar(false),
	}
	if err := tx.Create(room).Error; err != nil {
		return nil, err
	}
	return ChatRoomResource{
		RoomID:        room.ID,
		Name:          in.Name,
		Type:          RTGroup,
		NumberOfUsers: IntVar(len(allRoomUsersIds)),
		UserIDs:       allRoomUsersIds,
		CreatedAt:     room.CreatedAt,
	}, nil
}

func GetRoomMessages(c *fiber.Ctx, u *User, roomID string) (*PaginatedData[SentMessageResource], error) {
	// validate roomID to be a valid UUID
	if _, err := UUIDFromString(roomID); err != nil {
		return nil, fiber.NewError(fiber.StatusUnprocessableEntity, "invalid room_id")
	}
	tx := DB()

	var count int64
	if err := tx.Model(&ChatRoom{}).
		Where("id = ?", roomID).
		Where("? = ANY(users_ids)", u.ID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, fiber.NewError(fiber.StatusUnauthorized,
			"room not found, or you are not a member of this room")
	}

	txx := tx.Where("chat_room_id = ?", roomID)

	o, err := Paginate(c, ChatMessage{}, txx, func(tx *gorm.DB) *gorm.DB {
		// return tx.Joins("CreatedBy", DB().Where(`"CreatedBy".id <> ?`, u.ID)).Order("created_at DESC")
		return tx.Joins("CreatedBy").Order("created_at DESC")
	})
	if err != nil {
		return nil, err
	}

	newO, err := TransformPaginatedData(o, func(data ChatMessage) (SentMessageResource, error) {
		isMyMessage := data.CreatedByID == u.ID
		return SentMessageResource{
			ID:        data.ID,
			Content:   data.Content,
			Type:      data.Type,
			SentAt:    data.CreatedAt,
			EditedAt:  data.EditedAt,
			MyMassage: isMyMessage,
			SenderID:  &data.CreatedByID,
			RoomID:    &data.ChatRoomID,
			User:      *data.CreatedBy,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return newO, nil
}

type SendMessageInput struct {
	U           *User
	OtherUserID string
	RoomID      string
	Content     string
}

func SendMessageSync(in *SendMessageInput) error {
	newMessageOut, err := createNewMessage(in)
	if err != nil {
		return err
	}
	room := newMessageOut.Room
	msg := newMessageOut.Message
	currentUserId := in.U.ID.String()
	usersIds := []string{currentUserId}
	for _, id := range room.UsersIDs {
		if id != currentUserId {
			usersIds = append(usersIds, id)
		}
	}
	BroadcastWSMassage(usersIds, func(userId string) WSClientEventMessage {
		isMyMessage := userId == currentUserId
		return WSClientEventMessage{
			Type: WSMessageEvent,
			DataModel: SentMessageResource{
				ID:        msg.ID,
				Content:   msg.Content,
				Type:      msg.Type,
				SentAt:    msg.CreatedAt,
				EditedAt:  msg.EditedAt,
				MyMassage: isMyMessage,
				SenderID:  &msg.CreatedByID,
				RoomID:    &msg.ChatRoomID,
				User:      *in.U,
			},
		}
	})
	return nil
}

type SSEType string

const (
	SSEMessageEvent SSEType = "message"
)

type SocketSentEvent struct {
	Event SSEType         `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type SSEMessage struct {
	OtherUserID string `json:"other_user_id"`
	ChatRoomID  string `json:"room_id"`
	Content     string `json:"content"`
}

func ReceiveWSEvent(currentUser *User, data []byte) error {
	var sse SocketSentEvent
	if err := json.Unmarshal(data, &sse); err != nil {
		return err
	}

	currentUserId := currentUser.ID.String()

	switch sse.Event {
	case SSEMessageEvent:
		var sseData SSEMessage
		if err := json.Unmarshal(sse.Data, &sseData); err != nil {
			return err
		}
		newMessageOut, err := createNewMessage(&SendMessageInput{
			U:           currentUser,
			OtherUserID: sseData.OtherUserID,
			RoomID:      sseData.ChatRoomID,
			Content:     sseData.Content,
		})
		if err != nil {
			return err
		}
		room := newMessageOut.Room
		msg := newMessageOut.Message

		usersIds := []string{currentUserId}
		for _, id := range room.UsersIDs {
			if id != currentUserId {
				usersIds = append(usersIds, id)
			}
		}

		if newMessageOut.NewRoomCreated {
			crType := RTGroup
			isPrivate := len(room.UsersIDs) == 2 && *room.PeerToPeer
			if isPrivate {
				crType = RTPrivate
			}
			BroadcastWSMassage(usersIds, func(userId string) WSClientEventMessage {
				isMyMessage := userId == currentUserId

				model := ChatRoomResource{
					RoomID:        room.ID,
					Type:          crType,
					Name:          room.Name,
					UserIDs:       room.UsersIDs,
					CreatedAt:     room.CreatedAt,
					NumberOfUsers: IntVar(len(room.UsersIDs)),
					LastMessage: &SentMessageResource{
						ID:        msg.ID,
						Content:   msg.Content,
						Type:      msg.Type,
						SentAt:    msg.CreatedAt,
						EditedAt:  msg.EditedAt,
						MyMassage: isMyMessage,
						SenderID:  &msg.CreatedByID,
						RoomID:    &msg.ChatRoomID,
						User:      *currentUser,
					},
				}
				if isPrivate {
					otherUser := newMessageOut.OtherUser

					if isMyMessage {
						model.Name = currentUser.Name
						model.OtherUser = otherUser
					} else {
						model.Name = otherUser.Name
						model.OtherUser = currentUser
					}
				}
				return WSClientEventMessage{
					Type:      WSNewRoomEvent,
					DataModel: model,
				}
			})
		}

		go func() {
			// delay the message broadcast if the room is new
			if newMessageOut.NewRoomCreated {
				<-time.After(500 * time.Millisecond)
			}

			BroadcastWSMassage(usersIds, func(userId string) WSClientEventMessage {
				isMyMessage := userId == currentUserId
				return WSClientEventMessage{
					Type: WSMessageEvent,
					DataModel: SentMessageResource{
						ID:        msg.ID,
						Content:   msg.Content,
						Type:      msg.Type,
						SentAt:    msg.CreatedAt,
						EditedAt:  msg.EditedAt,
						MyMassage: isMyMessage,
						SenderID:  &msg.CreatedByID,
						RoomID:    &msg.ChatRoomID,
						User:      *currentUser,
					},
				}
			})
		}()
	}

	return nil
}

type newMessageOutput struct {
	Room           ChatRoom
	OtherUser      *User // only for private chat
	Message        ChatMessage
	NewRoomCreated bool
}

func createNewMessage(in *SendMessageInput) (*newMessageOutput, error) {
	if in.U.ID.String() == in.OtherUserID {
		return nil, fiber.NewError(fiber.StatusBadRequest, "cannot send message to yourself")
	}
	if in.RoomID == "" && in.OtherUserID == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "room_id or other_user_id is required")
	}

	tx := DB()

	currentUser := in.U
	room := &ChatRoom{}
	otherUser := &User{}
	roomAlreadyExists := false

	if in.RoomID != "" {
		if err := tx.Where("id = ?", in.RoomID).
			Where("? = ANY(users_ids)", currentUser.ID).
			First(room).Error; err != nil &&
			!errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if room.ID.IsEmpty() {
			return nil, fiber.NewError(fiber.StatusBadRequest,
				"room not found, or you are not a member. If you messaging another user, Try sending message to the other user and not the room for the first time")
		}
		roomAlreadyExists = true
		otherUser = nil
	} else {
		if err := tx.Where("id = ?", in.OtherUserID).
			First(otherUser).Error; err != nil &&
			!errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if otherUser.ID.IsEmpty() {
			return nil, fiber.NewError(fiber.StatusBadRequest, "massaged user not found")
		}

		if err := tx.
			Where("users_ids::TEXT[] @> ARRAY[?, ?]", currentUser.ID, otherUser.ID).
			Where("array_length(users_ids, 1) = 2").
			Where("peer_to_peer = TRUE").
			First(room).Error; err != nil &&
			!errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		roomAlreadyExists = !room.ID.IsEmpty()
	}

	var msg *ChatMessage

	newChatRoomCreated := false

	txError := tx.Transaction(func(tx *gorm.DB) error {
		roomId := room.ID
		if roomId.IsEmpty() {
			roomId = NewUUIDv4()
		}
		msg = &ChatMessage{
			ChatRoomID:  roomId,
			CreatedByID: currentUser.ID,
			Content:     in.Content,
			Type:        CMTypeText,
		}
		if roomAlreadyExists {
			if err := tx.Create(msg).Error; err != nil {
				return err
			}
			room.LatestMessageID = &msg.ID
			if err := tx.Updates(room).Error; err != nil {
				return err
			}
		} else {
			// create a new room
			room = &ChatRoom{
				ID:         roomId,
				UsersIDs:   StringArray{currentUser.ID.String(), in.OtherUserID},
				Name:       "private chat between " + currentUser.Name + " and " + otherUser.Name,
				UsersLimit: 2,
				PeerToPeer: BoolVar(true),
			}
			if err := tx.Create(room).Error; err != nil {
				return err
			}
			msg.ChatRoomID = room.ID
			if err := tx.Create(msg).Error; err != nil {
				return err
			}
			newChatRoomCreated = true
			room.LatestMessageID = &msg.ID
			if err := tx.Updates(room).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if txError != nil {
		return nil, txError
	}

	return &newMessageOutput{
		Room:           *room,
		Message:        *msg,
		OtherUser:      otherUser,
		NewRoomCreated: newChatRoomCreated,
	}, nil
}
