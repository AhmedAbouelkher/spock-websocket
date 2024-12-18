# Spock Websocket

A simple websocket server that can be used to send messages to clients in real time. The project consists of a websocket server written in Go, a web client written in HTML, CSS, and Javascript, and a mobile app written in Flutter.

Web Client:

![web client demo](./demos/web_client_demo.gif)

> [!NOTE]
> User by default is a member of all his rooms and is listening to all messages from all rooms. User can only send messages to rooms he is a member of.

Mobile App:

![mobile app demo](./demos/mobile_app_demo.gif)

## Table of Contents

- [Spock Websocket](#spock-websocket)
  - [Motivation](#motivation)
  - [Table of Contents](#table-of-contents)
  - [TODO](#todo)
  - [Pre-requisites](#pre-requisites)
  - [Project Structure](#project-structure)
  - [Running the web server](#running-the-web-server)
    - [Server Environment Variables](#server-environment-variables)
  - [Running the web client](#running-the-web-client)
    - [Web Client Command Line Arguments](#web-client-command-line-arguments)
    - [Database Schema](#database-schema)
  - [Running the mobile app (Flutter)](#running-the-mobile-app-flutter)
    - [Prerequisites for mobile app](#prerequisites-for-mobile-app)
    - [Prerequisites for Running on iOS](#prerequisites-for-running-on-ios)
  - [Credits](#credits)
  - [Contributing](#contributing)
  - [License](#license)

## Motivation

This project was created to learn how to create a websocket server in Go almost from scratch, and aims to showcase real life usage of websockets in both the server and client side.

## TODO

- [X] Build a simple websocket server that can connect to multiple clients with the same user id.
- [X] Build a simple web client and mobile app that can connect to the websocket server.
- [X] Build a simple database to store the messages.
- [X] One to one chat functionality. (Two users can chat with each other)
- [X] Group chat functionality. (Multiple users can chat in the same chat room)
- [ ] Expand the project to handle more advanced features in the web client.
- [ ] Improve chat functionality by adding more features like sending images, videos, etc.
- [ ] Improve chat functionality by expanding the chat to multiple users. (Group chat)
- [ ] Provide more comprehensive demo of the project on the web and mobile side.
- [ ] Build an Android native app and an iOS native app to connect to the websocket server.
- [ ] Implement end-to-end encryption for the messages.
- [ ] Users can delete and edit their chat messages.

## Pre-requisites

You need to have minimum understanding of the following:

- Go
- Websockets
- Javascript, HTML, CSS (specifically Tailwind CSS) _optional_
- Flutter, Dart (for the mobile app side) _optional_
- PostgreSQL (for the database)
- Goose for database migrations
- Docker Compose (for running the PostgreSQL database) _optional_

## Project Structure

```bash
.
├── flutter_app
├── server
└── webclient
```

- `flutter_app`: Contains the mobile app that connects to the websocket server.
- `server`: Contains the websocket server written in Go.
- `webclient`: Contains the web client that connects to the websocket server.

## Running the web server

The socket/http server is completely written in Go. To run the server, you need to have Go installed on your machine.

```bash
# clone the repository first
cd server
go mod tidy # to install the dependencies
go run main.go # a local socket/http server will start on port 3000 by default
```

or for to continuously run the server when you make changes, you can use this [Makefile](server/Makefile) to run the server:

```bash
# clone the repository first
cd server
go mod tidy # to install the dependencies
make watch # a local socket/http server will start on port 3000 by default
```

### Server Environment Variables

You can set the following environment variables to configure the server:

- `PORT`: The port on which the server will run. Default is `3000`.
- `JWT_SECRET`: The secret key to sign the JWT tokens. No default value.
- `ENV`: The environment in which the server is running. Default is `local`. Values can be `local`, or any other value.
- `POSTGRESQL_URL`: The URL to the PostgreSQL database. No default value. e.g. `postgresql://postgres:default@localhost:6523/spock?sslmode=disable`.

> [!TIP]
> Checkout this [docker-compose.yml](./server/docker-compose.yml) file to see how to run a PostgreSQL database locally in a docker container for development.

### Database Schema

You can find the database schema in the [db_schema.dbml](./server/db_schema.dbml) file. The schema is as follows:
![spock db schema](./demos/spock_db_schema.png)

The SQL migration files can be found in the [migrations](./server/migrations) folder. The migrations are written in SQL and are managed by [Goose](https://github.com/pressly/goose) so you need to have Goose installed on your machine to run the migrations.

```bash
# clone the repository first
cd server
go mod tidy # to install the dependencies
go get -u github.com/pressly/goose/cmd/goose # to install goose, or use `brew install goose` on macOS
goose -dir ./migrations postgres "YOUR_POSTGRESQL_URL" up # to run the migrations
```

## Running the web client

> [!IMPORTANT]
> Make sure you have the [webserver server](#running-the-web-server) running before running the web client. The web client connects to the websocket server.

The web client is a simple HTML, CSS, and Javascript application that connects to the websocket server. To run the web client, you need to have a browser installed on your machine.

```bash
# clone the repository first
cd webclient
go run main.go --api-url='API_URL' --ws-url='SERVER_WS_URL' # a local web server will start on port 3333 by default
```

### Web Client Command Line Arguments

- `port`: The port on which the web server will run. Default is `3333`.
- `api-url`: The URL to the [API server](#running-the-web-server). No default value. e.g. `http://localhost:4444/api/v1`.
- `ws-url`: The URL to the [websocket server]((#running-the-web-server)). No default value. e.g. `ws://localhost:4444/api/v1/ws/chat`.

> [!NOTE]
> In case of updating the html, css, or javascript files, you do not need to restart the server. The server will automatically fetch the newest files.

## Running the mobile app (Flutter)

> [!IMPORTANT]
> Make sure you have the [webserver server](#running-the-web-server) running before running the web client. The web client connects to the websocket server.

The mobile app is a simple Flutter application that connects to the websocket server. To run the mobile app, you need to have Flutter installed on your machine.

### Prerequisites for mobile app

- Make sure you have the Flutter SDK installed on your machine. You can follow the instructions [here](https://flutter.dev/docs/get-started/install).
- I am using Flutter version `3.24.0` and Dart version `3.5.0` as of the time of writing this README. You can check your Flutter and Dart versions by running `flutter --version` in your terminal.
- Make sure you have an emulator or a physical device connected to your machine.

    ```bash
    # clone the repository first
    cd flutter_app
    flutter pub get # to install the dependencies
    flutter run # a local mobile app will start on your emulator or physical device
    ```

- You can update the `API_URL` or `WS_URL` in the [lib/utils.dart](./flutter_app/lib/utils.dart) file to match the URL of the [webserver server](#running-the-web-server).

### Prerequisites for Running on iOS

- Make sure you have Xcode installed on your machine. I am using Xcode version `16.2 (16C5031c)` as of the time of writing this README.
- You can easily run the app on an iOS simulator by running `flutter run` in the terminal.
- I recommend manually installing pods by running `cd ios && pod install --repo-update ; cd ..` in the terminal before running the app.

## Credits

- [Star Trek](https://www.startrek.com/) for the the amazing series (_except for Discovery we do not talk about that_).
- [gorilla/websocket](https://github.com/gorilla/websocket) for the websocket server.
- [gofiber/fiber](https://github.com/gofiber/fiber) for the web server.
- [Tailwind CSS](https://tailwindcss.com/) for the web client.
- [Claude.ai](https://claude.ai/) for helping me building the web client.
- [PostgreSQL](https://www.postgresql.org/) for the database.
- [pressly/goose](https://github.com/pressly/goose) for the database migrations.
- [dbdiagram.io](https://dbdiagram.io/) for the database schema.
- [Docker Compose](https://docs.docker.com/compose/) for the orchestration.
- [Flutter](https://flutter.dev/) for the mobile app.
- [Dicebear](https://www.dicebear.com/) for the avatars.
- [lorenzodonini/ocpp-go](https://github.com/lorenzodonini/ocpp-go) for giving me the idea to build a websocket server, and frankly, some code snippets.
- [web_socket_client](https://pub.dev/packages/web_socket_client) for giving me inspiration on how to connect to a websocket server in Flutter.

## Contributing

If you have any suggestions, improvements, or issues, feel free to open an issue or a pull request. I am always open to new ideas and improvements.

## License

`Spock` is licensed under the MIT License. See [LICENSE](./LICENSE) for the full license text. This mean that this code is free to use, modify, distribute, and use for commercial purposes. So feel free to use it in your own projects, or modify it to suit your needs.
