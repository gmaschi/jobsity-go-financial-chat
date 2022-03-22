# jobsity-go-financial-chat

## Setting up the application

For convenience, the .env file is present in the repo.

To set up the application, run the following commands from the Makefile

#### Database

- create database: make postgres-up
- init database schema: make postgres-migrateup

#### RabbitMQ

- start broker: make rabbitmq-up

#### Application

- start the application: make server

## Using the chat

With everything set up and the server running, to use the application, one should go to localhost:8080/users/login.

There are two options at this page, login and sign up. When creating a user or signing in, if the request is successful, the user will be redirected to the chat lobby page at /chat.

### Creating a user

To create a user, there are two rules:

1. username must contain only alphanumeric characters;
2. password must have at least 6 characters.

### Chatting

At /chat, the user will be able to select from a set of different rooms to chat. There could be an unlimited number of rooms.

When selecting a room, the user will enter that room at /chat/:roomID and will be able to chat with other users.

If a non-authenticated user tries to join the chat lobby or a specific chat room, he/she will be redirected to the login/sign-up page.

## Running Tests

To run the application tests, both the message broker and the database containers must be up and running.

With everything set up, run "make test".

## Cleaning up

To clean everything up and remove database and message broker containers, just run "make clean-all".