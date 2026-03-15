# URL Shortener Microservice

A lightweight URL shortener microservice written in **Go**, built as a RESTful API for managing short links.

---

## Motivation

This project was created as a learning exercise while exploring **Go** and the **Clean Architecture** principles.  
It follows Go project layout best practices and demonstrates how to structure a microservice that is flexible, testable, and easy to extend.

---

## Features

- **Clean Architecture** separation of layers:
  - **Repository** – data access layer
  - **Use Cases** – business logic
  - **Entities & Models**
  - **Delivery Layer** – HTTP, gRPC, and CLI transports over the same use cases
- **Health check endpoint** – useful for Kubernetes liveness/readiness probes
- **OpenAPI documentation** – REST API described via a YAML spec
- **gRPC contract** – protobuf-defined API for short-link and health operations
- **Docker Compose** & [manage.sh] script – simple project startup
- **Graceful shutdown** – stops accepting new HTTP and gRPC connections and allows in-flight calls to finish during rolling updates
- **No web frameworks used** – HTTP stack uses Go’s standard library, while gRPC uses the official `grpc-go` package

---

## Clean Architecture Approach

The project applies Clean Architecture principles as strictly as possible:

- **Repository Layer**
  - Exposes only CRUD methods
  - Does not reveal database implementation details
  - Database can be swapped with minimal effort

- **Use Cases**
  - Contain only business logic
  - Do not know anything about the database (interact only through repository interfaces)
  - Easily testable and mockable

- **Delivery Layer**
  - Independent from business logic
  - Translates domain models into **DTOs** (Data Transfer Objects)
  - Never exposes domain models directly
  - Implemented as REST API, gRPC, and CLI using the same usecase contracts

This strict separation allows changing the transport layer, database, or other dependencies with minimal effort.

---

## Notes on Implementation

- Many GitHub examples labeled as _clean architecture_ break its core rules (e.g., exposing business models directly in transport, or mixing errors across layers).  
  This project aims to be a stricter example, based on my current understanding.
- Writing in Clean Architecture takes more effort initially, but results in:
  - Easier testing
  - Flexible infrastructure changes
  - Stronger decoupling between layers
- This is my first project using Clean Architecture, so feedback and improvement suggestions are welcome.

---

## Future Improvements

- Add proper test coverage
- Improve error handling (e.g., with error codes or custom error types)

---

## Dependencies

- **MongoDB**
- **Redis**

---

## Running the Project

You can start the project easily using the provided [manage.sh] script:

```bash
# start the infrastructure
./manage.sh dc:up

# ensure the .env file is present before running the application
./manage.sh run

# regenerate protobuf and gRPC generated files (after editing .proto)
./manage.sh proto
```

### CLI Mode

The binary now behaves as a CLI entrypoint.

- With no command, it starts the HTTP server (same behavior as before).
- With a command, it runs the CLI action.

Examples:

```bash
# create a short URL
./manage.sh run short create --url https://example.com/path

# expand by key (pure lookup, does not increment hits)
./manage.sh run short expand --key abc123

# delete by key
./manage.sh run short delete --key abc123
```

---

## License

This project is released under the [MIT license].

[MIT license]: LICENSE
[manage.sh]: manage.sh
