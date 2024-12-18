package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func SetupRouters(app *fiber.App) {
	app.All("/ping", func(c *fiber.Ctx) error { return c.SendString("pong") })

	apiV1 := app.Group("/api/v1")
	// auth apis
	{
		authApis := apiV1.Group("/auth")

		authApis.Post("/login", handleLogin)
		authApis.Post("/register", handleRegister)

		// create me path with auth middleware
		authApis.Get("/me", AuthMiddleware(), handleMe)
	}
	// chat apis
	{
		chatApis := apiV1.Group("/chat")
		chatApis.Get("/discover-users", AuthMiddleware(), handleDiscoverUsers)
		chatApis.Get("/discover-rooms", AuthMiddleware(), handleDiscoverRooms)
		chatApis.Get("/rooms", AuthMiddleware(), handleGetRooms)
		chatApis.Post("/create-group-room", AuthMiddleware(), handleCreateGroupRoom)
		chatApis.Get("/room-messages/:room_id", AuthMiddleware(), handleRoomMessages)
		chatApis.Post("/send-message-sync", AuthMiddleware(), handleSendMessage)
	}
	// websockets
	apiV1.Get("/ws/chat", WSAuthMiddleware(), adaptor.HTTPHandlerFunc(HandleChatWS))
}

func handleLogin(c *fiber.Ctx) error {
	type P struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	var payload P
	if err := ParseAndValidate(c, &payload); err != nil {
		return err
	}
	out, err := LoginUser(&LoginUserInput{
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		return err
	}
	return c.JSON(out)

}

func handleRegister(c *fiber.Ctx) error {
	type P struct {
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	var payload P
	if err := ParseAndValidate(c, &payload); err != nil {
		return err
	}
	out, err := RegisterNewUser(&RegisterNewUserInput{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		return err
	}
	return c.JSON(out)
}

func handleMe(c *fiber.Ctx) error {
	// get user from locals
	user, ok := c.Locals("user").(*User)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "user not found")
	}
	return c.JSON(user)
}

func handleDiscoverUsers(c *fiber.Ctx) error {
	// get user from locals
	user, ok := c.Locals("user").(*User)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "user not found")
	}
	out, err := DiscoverUsers(c, user)
	if err != nil {
		return err
	}
	return c.JSON(out)
}

func handleDiscoverRooms(c *fiber.Ctx) error {
	// get user from locals
	user, ok := c.Locals("user").(*User)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "user not found")
	}
	out, err := DiscoverRooms(c, user)
	if err != nil {
		return err
	}
	return c.JSON(out)
}

func handleGetRooms(c *fiber.Ctx) error {
	// get user from locals
	user, ok := c.Locals("user").(*User)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "user not found")
	}
	out, err := GetRoomsByUserID(c, user)
	if err != nil {
		return err
	}
	return c.JSON(out)
}

func handleCreateGroupRoom(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*User)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "user not found")
	}
	type P struct {
		Name          string   `json:"name" validate:"required,max=255"`
		OtherUsersIDs []string `json:"other_users_ids" validate:"required,min=2,unique,dive,uuid"`
	}
	var payload P
	if err := ParseAndValidate(c, &payload); err != nil {
		return err
	}
	out, err := CreateGroupRoom(&CreateGroupRoomInput{
		U:             user,
		Name:          payload.Name,
		OtherUsersIDs: payload.OtherUsersIDs,
	})
	if err != nil {
		return err
	}
	return c.JSON(out)
}

func handleRoomMessages(c *fiber.Ctx) error {
	// get user from locals
	user, ok := c.Locals("user").(*User)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "user not found")
	}
	roomID := c.Params("room_id")
	if roomID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "room_id is required")
	}
	out, err := GetRoomMessages(c, user, roomID)
	if err != nil {
		return err
	}
	return c.JSON(out)
}

func handleSendMessage(c *fiber.Ctx) error {
	// get user from locals
	user, ok := c.Locals("user").(*User)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "user not found")
	}
	type P struct {
		OtherUserID string `json:"other_user_id" validate:"omitempty,uuid"`
		RoomID      string `json:"room_id" validate:"omitempty,uuid"`
		Content     string `json:"content" validate:"required,max=255"`
	}
	var payload P
	if err := ParseAndValidate(c, &payload); err != nil {
		return err
	}
	err := SendMessageSync(&SendMessageInput{
		U:           user,
		OtherUserID: payload.OtherUserID,
		RoomID:      payload.RoomID,
		Content:     payload.Content,
	})
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "message sent"})
}
