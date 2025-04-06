# URL Shortener
A simple and efficient URL shortener service written in Go, using the Gin web framework, Redis for caching, and SQLite3 for persistent storage.

## Prerequisites
- Go 1.21+
- Redis server with Bloom Filter module

## Setup
1. Install Redis.
- **Option 1**: Install [Redis Stack](https://redis.io/docs/latest/operate/oss_and_stack/install/install-stack/) which contains Bloom Filter module and other new features.

- **Option 2**: Install [Redis server](https://redis.io/docs/latest/get-started/) and [Bloom Filter module](https://github.com/RedisBloom/RedisBloom)

2. Clone the respository
```
git clone https://github.com/csBenClarkson/url-shortener.
cd url-shortener
```

3. Get Go dependencies
```
go mod tidy
```

4. Run the server with default settings
```
go run .
```

## Usage
You can configurate the Redis server address, sqlite3 database file location, host address, etc. through program arguments. Use `go run . -h` to see the help page.

You can also build the binary and run it.
```
go build .
./url-shortener [...argument]
```

## License
This project use the MIT license. However, there are third-party modules used in this project. Check THIRD_PARTY_LICENSES for details.
