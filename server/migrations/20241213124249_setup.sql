-- +goose Up
-- +goose StatementBegin

CREATE TABLE "users" (
  "id" UUID PRIMARY KEY,
  "name" VARCHAR(255) NOT NULL,
  "profile_image_icon" TEXT,
  "email" VARCHAR(255) UNIQUE NOT NULL,
  "password" VARCHAR(255) NOT NULL,
  "created_at" TIMESTAMP(0) NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" TIMESTAMP(0) NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "deleted_at" TIMESTAMP(0)
);

CREATE TABLE "chat_rooms" (
  "id" UUID PRIMARY KEY,
  "name" VARCHAR(255) NOT NULL,
  "users_limit" INTEGER NOT NULL,
  "users_ids" VARCHAR(255)[] NOT NULL,
  "peer_to_peer" boolean DEFAULT true,
  "last_message_content" text,
  "last_message_type" VARCHAR(255),
  "last_message_sent_at" TIMESTAMP(0),
  "created_at" TIMESTAMP(0) NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" TIMESTAMP(0) NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "deleted_at" TIMESTAMP(0)
);

CREATE TABLE "chat_messages" (
  "id" SERIAL PRIMARY KEY,
  "chat_room_id" UUID NOT NULL,
  "created_by_id" UUID NOT NULL,
  "content" TEXT NOT NULL,
  "type" VARCHAR(255) NOT NULL,
  "created_at" TIMESTAMP(0) NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "edited_at" TIMESTAMP(0),
  "deleted_at" TIMESTAMP(0)
);

ALTER TABLE "chat_messages" ADD FOREIGN KEY ("chat_room_id") REFERENCES "chat_rooms" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "chat_messages" ADD FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE "chat_messages";
DROP TABLE "chat_rooms";
DROP TABLE "users";

-- +goose StatementEnd
