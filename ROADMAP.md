# Build Your Own Redis — Roadmap

Learning Go by re-implementing Redis's core. Style: I explain + spec each stage,
you write the code, I review. Check off stages as we complete them.

## Stage 1 — Core essentials
- [x] 1.1 TCP server: bind to a port, accept one connection, read raw bytes
- [x] 1.2 RESP protocol: parse a PING, reply +PONG\r\n
- [x] 1.3 Handle multiple concurrent clients (goroutines) + multi-command per connection
- [x] 1.4 ECHO command (done ahead of schedule alongside 1.2)
- [x] 1.5 In-memory store: SET / GET (map + mutex)
- [x] 1.6 DEL, EXISTS
- [x] 1.7a Key expiry: SET ... PX/EX, passive expiration (checked on GET/EXISTS)
- [x] 1.7b Active expiration (background sweeper goroutine, verified via debug log)

## Stage 2 — Persistence
- [ ] 2.1 RDB file format basics: read an existing RDB on startup
- [ ] 2.2 CONFIG GET/SET (dir, dbfilename)
- [ ] 2.3 KEYS command using loaded RDB data

## Stage 3 — Replication
- [ ] 3.1 Master/replica handshake (REPLCONF, PSYNC)
- [ ] 3.2 Propagate writes from master to replicas
- [ ] 3.3 WAIT command

## Stage 4 — Extras
- [ ] 4.1 Pub/Sub: SUBSCRIBE, PUBLISH
- [ ] 4.2 Transactions: MULTI, EXEC, DISCARD
- [ ] 4.3 Streams (XADD, XRANGE) — stretch goal

## Notes / decisions log
(running log of design choices we make, so future sessions have context)
- Dev/test port is 6390, not 6379 (real Redis was already occupying 6379 on this machine).
- On this Windows machine, `net.Listen("tcp", ":6390")` binds IPv6 (`::`) only, not
  IPv4 `127.0.0.1` — `redis-cli -p 6390 ping` (which defaults to 127.0.0.1) gets
  "connection refused". Use `redis-cli -h localhost -p 6390 ...` instead. Not fixed
  in code (not worth solving yet); revisit if it becomes a real annoyance.
