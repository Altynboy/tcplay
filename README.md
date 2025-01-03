# tcplay: A Minimalist TCP Library for Go

GoTCP is a lightweight TCP library built for learning and understanding TCP communication in Go. It provides essential features for creating TCP servers and clients, with support for connection management, message handling, and graceful shutdown.

## Core Features

1. Connection Management

   - Server:

     - Start a TCP server.
     - Accept incoming client connections.

   - Client:

     - Connect to a TCP server.
     - Reconnect logic for connection failures.

2. Data Transmission

   - Support for:

     - Sending and receiving messages (byte streams).
     - Handling partial reads/writes.
     - Reading messages with delimiters or fixed length.

3. Graceful Shutdown

   - Ability to stop the server and clear all connections cleanly.
   - Handle closing clients connection without data loss or corruption.

4. Error Handling

   - Meaningful error messages for debugging.
   - Handle timeouts and unexpected disconnections.
