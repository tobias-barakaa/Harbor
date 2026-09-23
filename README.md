# Harbor

Harbor is a deployment platform being built from scratch in Go.

The goal is to understand and build the systems behind application deployment — from inspecting source code to eventually building, running, and deploying applications to remote servers.

> **Status:** Early development

## What Harbor Does Right Now

Harbor can inspect an application archive and identify useful information about the applications inside it.

For example:

```bash
go run ./cmd/harbor inspect ~/Downloads/frontend-backend.zip
```

The inspection process can detect things such as:

* `package.json` → Node.js application
* `requirements.txt` → Python application
* Possible application ports
* Detected runtime
* Possible deployment method

Example output:

```text
Application
────────────
Name: backend-js

Detected:
  ✓ package.json

Runtime:
  Node.js

Possible port:
  3000

Deployment method:
  Standard
```

## Project Direction

Harbor is being developed incrementally.

The intended pipeline is:

```text
Source Code
     │
     ▼
  Inspect
     │
     ▼
  Detect
     │
     ▼
  Plan
     │
     ▼
   Build
     │
     ▼
   Run
     │
     ▼
  Deploy
     │
     ▼
 Remote Server
```

Eventually, Harbor is intended to handle more of this lifecycle:

* Application inspection
* Runtime detection
* Deployment planning
* Application builds
* Process management
* Networking
* Containerization
* Remote server deployment
* Deployment status and logs
* Application management

## Why Build Harbor?

Rather than starting with an existing deployment platform and treating it as a black box, Harbor is an attempt to understand what happens underneath.

The project is being built piece by piece, starting with the fundamentals:

```text
Archive
  ↓
Filesystem inspection
  ↓
Application detection
  ↓
Deployment planning
  ↓
Build system
  ↓
Process/container execution
  ↓
Networking
  ↓
Remote deployment
```

## Tech Stack

* **Go**
* Linux
* Git
* Docker *(planned/integrated as the project evolves)*

## Project Structure

```text
harbor/
├── cmd/
│   └── ...
├── internal/
│   ├── inspect/
│   └── report/
├── go.mod
└── README.md
```

The project is intentionally being developed as a set of small components rather than one large application.

## Development

Clone the repository:

```bash
git clone https://github.com/tobias-barakaa/Harbor.git
cd Harbor
```

Run the project:

```bash
go run ./cmd/...
```

> The CLI and commands are currently under active development and may change as Harbor evolves.

## Roadmap

* [x] Initialize Go project
* [x] Git repository
* [x] Application archive inspection
* [x] Basic runtime detection
* [x] Basic application reporting
* [ ] Improved application detection
* [ ] Deployment planning
* [ ] Build system
* [ ] Process execution
* [ ] Application networking
* [ ] Container support
* [ ] Remote server management
* [ ] Remote deployment
* [ ] Deployment logs and status
* [ ] Desktop client

## Philosophy

Harbor is being built with a focus on understanding the underlying systems rather than simply wrapping existing tools.

The project starts with a simple question:

> **Given some application source code, how can a system understand what it is and determine how it should be deployed?**

That question will drive the project from inspection all the way to deployment.

## License

License to be determined.
