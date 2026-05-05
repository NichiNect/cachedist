# SKILLS.md — Go Distributed Cache Project

> Dokumen ini adalah **kontrak teknis** untuk AI assistant (vibecoding). Setiap sesi build harus dimulai dengan membaca file ini. Jangan mengubah keputusan arsitektur yang sudah ditetapkan tanpa alasan eksplisit.

***

## Identitas Project

- **Nama project**: `cachedist`
- **Bahasa**: Go 1.22+
- **Tujuan**: Membangun distributed in-memory cache mirip Redis dari nol untuk keperluan portfolio backend developer dan study-case
- **Repo structure**: Monorepo, semua komponen dalam satu repository

***

## Prinsip Umum

1. **Build incrementally** — setiap step harus menghasilkan program yang bisa dijalankan dan diuji
2. **No premature optimization** — tulis yang benar dulu, optimize belakangan
3. **Explicit over implicit** — tidak ada magic, semua konfigurasi eksplisit
4. **Test setiap step** — setiap step wajib punya unit test minimal untuk happy path
5. **Idiomatic Go** — ikuti Go convention: error handling eksplisit, interface kecil, composition over inheritance

***

## Arsitektur Target (Final)

```
cachedist/
├── cmd/
│   ├── server/          # Entry point cache server
│   │   └── main.go
│   └── cli/             # CLI tool untuk interaksi manual
│       └── main.go
├── internal/
│   ├── cache/           # Core cache engine (step 1-2)
│   │   ├── cache.go     # Interface & public API
│   │   ├── shard.go     # Shard implementation
│   │   ├── lru.go       # LRU eviction linked list
│   │   ├── item.go      # Item struct + TTL
│   │   └── cache_test.go
│   ├── server/          # HTTP server (step 1)
│   │   ├── server.go
│   │   ├── handler.go
│   │   └── server_test.go
│   ├── cluster/         # Cluster & node management (step 3-5)
│   │   ├── node.go      # Node struct & metadata
│   │   ├── ring.go      # Consistent hash ring
│   │   ├── ring_test.go
│   │   └── manager.go   # Node discovery & health
│   ├── replication/     # Write replication (step 4)
│   │   ├── replicator.go
│   │   └── replicator_test.go
│   ├── grpc/            # gRPC server & client (step 3+)
│   │   ├── server.go
│   │   └── client.go
│   └── metrics/         # Prometheus metrics (step 6)
│       └── metrics.go
├── proto/
│   └── cache.proto      # Protobuf definitions
├── config/
│   └── config.go        # Config dari environment / YAML
├── sdk/                 # Client SDK (step 3)
│   └── client.go        # Consistent hashing + routing
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
├── Makefile
├── go.mod
├── go.sum
├── README.md
└── SKILLS.md            # File ini
```

***

## Keputusan Teknis (JANGAN DIUBAH)

### Storage Engine
- Menggunakan native Go `map[string]*Item` per shard
- **Tidak** menggunakan library cache eksternal (bigcache, freecache, dll)
- Semua harus diimplementasi sendiri untuk tujuan pembelajaran

### Concurrency Model
- **256 shard** per node (configurable via `CACHE_NUM_SHARDS`)
- Setiap shard punya `sync.RWMutex` sendiri
- `RLock` untuk GET, `Lock` untuk SET/DELETE
- Background goroutine untuk TTL cleanup (interval: 30 detik, configurable)

### Eviction Policy
- **LRU** sebagai default (Least Recently Used)
- Implementasi: `doubly linked list` + `hashmap` untuk O(1) get dan evict
- Max items per node: configurable via `CACHE_MAX_ITEMS` (default: 1,000,000)

### Hashing
- **FNV-1a 32-bit** untuk shard selection (sudah ada di Go stdlib: `hash/fnv`)
- **Consistent hashing** untuk node selection di SDK (implementasi sendiri dengan virtual nodes)
- Virtual nodes per physical node: 150 (configurable)

### Communication Protocol
- **HTTP/JSON** untuk client-facing API (step 1-2, agar mudah di-test dengan curl)
- **gRPC + Protobuf** untuk inter-node communication (step 3+)
- Port convention: HTTP `:700{N}`, gRPC `:800{N}` (N = node index)

### Replication
- **Replication factor**: 2 (setiap key ada di 2 node)
- **Write quorum**: 2 (primary + 1 replica harus ACK)
- **Read**: dari primary node saja (tidak read repair untuk simplisitas)
- Replikasi bersifat **synchronous** untuk write, background untuk catch-up setelah node recovery

### Health Check
- Heartbeat interval: 5 detik
- Node dianggap dead setelah: 3x miss berturut-turut
- Dead node dikeluarkan dari hash ring secara otomatis

### TTL
- Granularity: detik (bukan millisecond)
- Lazy deletion: saat GET, cek expiry sebelum return
- Active deletion: background goroutine scan setiap 30 detik

***

## API Contract (HTTP)

Semua response menggunakan JSON dengan format:
```json
{ "success": true, "data": "...", "error": "" }
```

| Method | Endpoint | Body | Deskripsi |
|--------|----------|------|-----------|
| GET | `/get?key={key}` | - | Ambil value |
| POST | `/set` | `{"key":"...","value":"...","ttl":60}` | Set value (ttl dalam detik, 0 = no expiry) |
| DELETE | `/delete?key={key}` | - | Hapus key |
| GET | `/stats` | - | Info node: hit rate, item count, memory |
| GET | `/health` | - | Health check untuk heartbeat |
| GET | `/keys` | - | List semua key (debug only) |

***

## Environment Variables

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `CACHE_NODE_ID` | `node-1` | ID unik node |
| `CACHE_HTTP_PORT` | `7001` | Port HTTP server |
| `CACHE_GRPC_PORT` | `8001` | Port gRPC server |
| `CACHE_NUM_SHARDS` | `256` | Jumlah shard |
| `CACHE_MAX_ITEMS` | `1000000` | Max item per node |
| `CACHE_TTL_CLEANUP` | `30` | Interval cleanup TTL (detik) |
| `CACHE_PEERS` | `` | Comma-separated list peers: `host:grpcport` |
| `CACHE_REPLICATION_FACTOR` | `2` | Jumlah replika per key |
| `CACHE_VIRTUAL_NODES` | `150` | Virtual nodes per physical node |

***

## Go Modules & Dependencies

```
module github.com/NichiNect/cachedist

go 1.22

dependencies yang DIIZINKAN:
- google.golang.org/grpc          # gRPC
- google.golang.org/protobuf      # Protobuf
- github.com/prometheus/client_golang  # Metrics (step 6 saja)
- gopkg.in/yaml.v3                # Config file parsing
```

**Tidak diizinkan** menambahkan dependency lain tanpa alasan eksplisit yang dicatat di sini.

***

## Testing Convention

- File test: `*_test.go` di package yang sama
- Nama fungsi: `TestFunctionName_Scenario` (contoh: `TestCache_SetAndGet`, `TestLRU_EvictsLeastRecent`)
- Wajib ada benchmark: `BenchmarkCache_Get`, `BenchmarkCache_Set`
- Gunakan `t.Parallel()` untuk test yang bisa dijalankan paralel
- **Tidak** menggunakan testing library eksternal (cukup stdlib `testing`)

***

## Makefile Commands

```makefile
make run-node1    # Jalankan node 1 (port 7001/8001)
make run-node2    # Jalankan node 2 (port 7002/8002)
make run-node3    # Jalankan node 3 (port 7003/8003)
make test         # Run semua test
make bench        # Run benchmarks
make proto        # Generate dari .proto file
make docker-up    # Jalankan 3 node via docker-compose
make docker-down  # Stop semua
make lint         # golangci-lint
```

***

## Step Overview

| Step | Fokus | Output yang Bisa Ditest |
|------|-------|------------------------|
| 1 | Single-node cache + HTTP API | `curl localhost:7001/set` berhasil |
| 2 | Sharding + LRU + TTL | Benchmark menunjukkan paralel shard lebih cepat |
| 3 | Multi-node + consistent hash SDK | SDK auto-route key ke node yang benar |
| 4 | gRPC replication | Key tetap ada setelah primary node dimatikan |
| 5 | Health check + auto-recovery | Node crash & rejoin tanpa data corruption |
| 6 | Prometheus metrics | Dashboard Grafana menampilkan hit rate |
| 7 | Benchmark vs Redis | Laporan perbandingan throughput ops/sec |

***

## Glosarium

| Term | Definisi dalam konteks project ini |
|------|------------------------------------|
| **Node** | Satu instance cache server |
| **Shard** | Partisi internal dalam satu node |
| **Primary** | Node yang bertanggung jawab untuk suatu key range |
| **Replica** | Node backup yang menyimpan salinan key dari primary |
| **Hash Ring** | Lingkaran virtual untuk distribusi key ke node |
| **Virtual Node** | Titik tambahan di hash ring untuk satu physical node |
| **Quorum** | Jumlah minimum node yang harus ACK sebelum operasi dianggap sukses |
| **TTL** | Time-to-live, waktu hidup sebuah cache entry |
| **Eviction** | Proses menghapus item dari cache saat kapasitas penuh |