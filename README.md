# goproxy 🚀

A high-performance, concurrent Layer-7 HTTP Load Balancer & Reverse Proxy built from scratch in pure Go.

`goproxy` receives incoming client HTTP requests, dynamically routes them across a pool of backend servers using Round-Robin load balancing, performs active background health checks, and executes automatic passive failover retries when backend connections drop.

---

## 🌟 Key Features

* **Round-Robin Load Balancing:** Evenly distributes incoming HTTP traffic across a pool of active backend targets.
* **Active Background Health Checks:** Periodically pings backend targets in a background goroutine loop using `time.Ticker`. Automatically evicts dead servers from rotation and restores them upon recovery.
* **Passive Failover & Seamless Retries:** Catches mid-flight connection drops using `httputil.ReverseProxy.ErrorHandler`, marks the failing backend dead, and transparently retries the client's request on another healthy server.
* **Thread-Safe Concurrency:** Built with `sync.Mutex` and `sync.RWMutex` protection. Fully verified for zero data races under heavy parallel loads (`go run -race .`).
* **Dynamic JSON Configuration:** Load proxy ports, health check intervals, and target backend URLs from `config.json` without recompiling binary source code.
* **Zero Dependencies:** Built 100% using the Go standard library (`net/http`, `net/http/httputil`, `sync`).

---

## 🏗️ Architecture Overview

```text
[ Client Request ]
       │
       ▼
┌──────────────┐
│   goproxy    │ ◄── [ Active Health Checker (ticker) ]
│  (Port 8080) │
└──────┬───────┘
       │ (Round-Robin Selection)
       ├─────────────────┬─────────────────┐
       ▼                 ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Backend 1   │  │  Backend 2   │  │  Backend 3   │
│ (Port 8081)  │  │ (Port 8082)  │  │ (Port 8083)  │
└──────────────┘  └──────────────┘  └──────────────┘
```

---

## 🚀 Quickstart

### 1. Prerequisites
Ensure you have **Go 1.22+** installed on your system.

### 2. Clone the Repository
```bash
git clone https://github.com/JejurkarYash/goproxy.git
cd goproxy
```

### 3. Run the Load Balancer
```bash
go run .
```

The server will load `config.json`, boot up the mock backend servers (`:8081`, `:8082`, `:8083`), start the background health checker, and listen for proxy requests on port `8080`.

### 4. Send Test Requests
Open another terminal and send HTTP requests to the load balancer:

```bash
for i in {1..6}; do curl http://localhost:8080/; echo; done
```

**Output:**
```text
Hello from Backend:8081
Hello from Backend:8082
Hello from Backend:8083
Hello from Backend:8081
Hello from Backend:8082
Hello from Backend:8083
```

---

## ⚙️ Configuration (`config.json`)

`goproxy` is configured via a simple `config.json` file in the project root:

```json
{
  "port": 8080,
  "backend_healthcheck_interval": 5,
  "backends": [
    "http://localhost:8081",
    "http://localhost:8082",
    "http://localhost:8083"
  ]
}
```

* `port`: The port `goproxy` listens on.
* `backend_healthcheck_interval`: Frequency (in seconds) of active background health pings.
* `backends`: Array of target backend server URLs.

---

## 🛡️ Race Detector Verification

To verify that `goproxy` is 100% free of data races under concurrent execution:

```bash
go run -race .
```

---

## 📁 Project Structure

```text
goproxy/
├── config.json    # JSON runtime configuration
├── config.go      # Configuration loader & struct models
├── backend.go     # Backend model, thread-safe health methods, & reverse proxy engine
├── pool.go        # ServerPool model, Round-Robin selection, & HealthChecker loop
└── main.go        # Main orchestrator, mock backends, & HTTP server
```

---

## 📝 License
Distributed under the MIT License.
