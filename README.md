# Simple DNS Server

This project is a basic DNS server implementation in Go, created as part of the CodeCrafters "Build Your Own X" challenge. It provides fundamental DNS functionality, including request decoding/encoding and request forwarding to other DNS servers.

## Features

- Basic DNS request decoding and encoding
- Request forwarding to other DNS servers

## Getting Started

## Prerequisites

This project requires Go to be installed on your system. If you don't have Go installed, you can download it from the official Go downloads page:

[https://go.dev/dl/](https://go.dev/dl/)

Choose the appropriate version for your operating system and follow the installation instructions provided on the Go website.

## Installation

After ensuring Go is installed on your system, follow these steps to set up Simple Git:

1. Clone the repository:
   ```
   git clone https://github.com/kakkarot9712/simple-dns-server-go
   ```

2. Navigate to the project directory:
   ```
   cd simple-dns-server-go
   ```

3. Build the project:
   ```
   go build -o mydnsserver ./app
   ```

This will create an executable named `mydnsserver` in your project directory.

### Usage

- To start the DNS server, run:
```
./mydnsserver
```
The server listens on port 2053.

- To start DNS server with reolver (Request forwarding), you can do so by passing `--resolver` flag along with server host and port like this:
```
./mydnsserver --resolver 127.0.0.1:2054
```

## Limitation

- This server can not resolve any DNS queries without resolver server.
- This server only reolves A records.
- This is very basic implimentation of DNS server with very limited features.

## Acknowledgments

- CodeCrafters for the "Build Your Own X" challenge