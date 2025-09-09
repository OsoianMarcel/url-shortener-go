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
  - **Delivery Layer** – currently only HTTP, with plans to add another transport to showcase the advantages of clean architecture
- **Health check endpoint** – useful for Kubernetes liveness/readiness probes
- **OpenAPI documentation** – REST API described via a YAML spec
- **Docker Compose** & [manage.sh] script – simple project startup
- **Graceful shutdown** – stops accepting new HTTP connections and allows existing requests to finish during rolling updates (avoiding `500` errors)
- **No frameworks used** – implemented entirely with Go’s standard library for learning purposes

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
  - Depend only on the methods required (not entire repository interfaces)
  - Easily testable and mockable

- **Delivery Layer**
  - Independent from business logic
  - Translates domain models into **DTOs** (Data Transfer Objects)
  - Never exposes domain models directly
  - Currently implemented as a REST API, but could be extended to CLI, gRPC, XML API, etc.

This strict separation allows changing the transport layer, database, or other dependencies with minimal effort.

---

## Notes on Implementation

- Many GitHub examples labeled as *clean architecture* break its core rules (e.g., exposing business models directly in transport, or mixing errors across layers).  
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
- Add an additional transport layer (most likely CLI)

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
```

---

## License

This project is released under the [MIT license].

[MIT license]: LICENSE
[manage.sh]: manage.sh
