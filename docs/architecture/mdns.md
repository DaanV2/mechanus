# mDNS — LAN Service Discovery

Mechanus implements a [Multicast DNS (RFC 6762)](https://datatracker.ietf.org/doc/html/rfc6762) server so clients on the same local network can discover the server automatically, without needing to know its IP address.

## How it works

On startup the server joins the mDNS multicast groups and begins responding to DNS-SD queries. When a browser or device queries for services of the configured type (default `_http._tcp.local`), the server responds with:

- a **PTR record** pointing to the service instance name
- an **SRV record** with the hostname and port
- a **TXT record** with optional metadata
- an **A record** mapping the hostname to the server's IPv4 address

This is the standard DNS-SD pattern, compatible with browsers and OS-level service browsers (e.g. Bonjour on macOS).

## Configuration

| Flag                 | Default            | Description                                     |
| -------------------- | ------------------ | ----------------------------------------------- |
| `--mdns.hostname`    | `mechanus`         | Hostname broadcast in SRV/A records             |
| `--mdns.servicetype` | `_http._tcp.local` | DNS-SD service type                             |
| `--mdns.ipv6`        | `false`            | Also join the IPv6 multicast group (`FF02::FB`) |

The port advertised is whatever port the HTTP server is listening on — it is passed through from the server config automatically.

## Network details

| Parameter              | Value         |
| ---------------------- | ------------- |
| UDP port               | 5353          |
| IPv4 multicast address | `224.0.0.251` |
| IPv6 multicast address | `FF02::FB`    |
| Record TTL             | 120 seconds   |

## Implementation

`infrastructure/transport/mdns.Server` binds to the multicast UDP address on startup and spawns a goroutine per connection (`serverConn`) to listen for incoming queries. Each query is matched against the configured service type; matching queries receive the PTR / SRV / TXT / A response bundle.

The mDNS server follows the same lifecycle as the rest of the application — `Server.Listen()` is called after initialization and `Server.Shutdown(ctx)` is called during graceful shutdown.

## Usage

When mDNS is active, any browser or device on the same LAN should be able to reach the server at:

```text
http://mechanus.local:<port>/
```

(assuming the default hostname). No manual IP configuration needed.
