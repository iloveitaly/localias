# Caddy 2.11.7 upgrade notes

Reviewed on October 26, 2026, upgrading from Caddy 2.10.0 (introduced in
[PR #62](https://github.com/peterldowns/localias/pull/62)) to the latest stable
release, 2.11.7.

## Benefits for localias

- **Streaming fixes.** 2.11.7 fixes an HTTP/2 proxy panic and premature stream
  termination introduced by 2.11.6's idle timeouts. Quiet SSE responses are no
  longer closed just because there is a pause between writes.
- **Incremental responses.** An upstream can return `Incremental: ?1` to request
  immediate forwarding. Caddy handles this automatically, so a development app
  can opt in without localias adding `flush_interval -1` to every proxy.
- **Proxy and connection improvements.** 2.11.6 improves partial-response
  flushing, propagates TCP half-close on upgraded streams, and fixes HTTP/3 on
  low-MTU links such as Tailscale. 2.11.7 also reduces TLS handshake overhead.
- **Security fixes.** The releases since 2.10.0 include fixes across routing,
  reverse proxying, TLS, and the admin API. Localias already disables the admin
  listener, but still benefits from the fixes in the embedded HTTP/TLS server.

Sources: [2.11.7 release notes](https://github.com/caddyserver/caddy/releases/tag/v2.11.7),
[2.11.6 release notes](https://github.com/caddyserver/caddy/releases/tag/v2.11.6),
[2.11.1 release notes](https://github.com/caddyserver/caddy/releases/tag/v2.11.1).

## Compatibility changes

| Change | Effect on localias |
| --- | --- |
| Caddy now requires Go 1.26 or newer. | The complete dependency update requires Go 1.26.9. CI reads `go.mod`, and the Nix lock now supplies a suitable compiler. |
| Request headers default to a 16 KiB limit, down from Go's previous 1 MiB default. | Apps with large cookies or tokens can receive HTTP 431. Caddy's `servers { max_header_size ... }` option can raise the limit if needed. |
| Reads of request bodies and response writes have one-minute idle timeouts. | A stalled transfer can be aborted; ordinary SSE pauses between writes do not count. Caddy supports `read_body_idle` and `write_idle` under `servers { timeouts { ... } }` for tuning. |
| Request headers containing underscores or dots are dropped by default. | Apps using such headers need explicit `expected_underscore_headers` or `expected_dot_headers` allowlists. These filters prevent header-name aliasing. |
| HTTPS upstreams automatically receive a Host header matching the upstream address. | Localias uses plain HTTP upstreams (`localhost:<port>`), so its original Host forwarding is unaffected. Keep this change in mind if HTTPS upstream support is added. |

The generated configuration keeps Caddy's new limits and header filtering. There
is currently no localias setting for these server options; the Caddy options
above describe possible future additions if a concrete application needs them.

Sources: [2.11.6 breaking changes](https://github.com/caddyserver/caddy/releases/tag/v2.11.6),
[Caddy server options](https://caddyserver.com/docs/caddyfile/options#servers),
[reverse proxy headers](https://caddyserver.com/docs/caddyfile/directives/reverse_proxy#headers).

## Configuration decisions

No new directives are needed for the upgrade. Keep internal, on-demand
certificates, the existing Localias CA identity and storage location, HTTP-only
aliases, and the disabled admin listener. The regression test adapts a mixed
configuration and checks that only HTTPS aliases receive on-demand internal
certificate policies, without an ask endpoint.

Other new features have limited use here: `tls_automate_names` manages names
without site blocks, while localias already creates a block for every alias;
ECH and DNS challenge settings target public DNS deployments; and Caddy's
`SIGUSR1` reload only applies to configurations initially loaded from a file by
its CLI. Localias loads its generated JSON through `caddy.Load`.

Sources: [2.11.1 features](https://github.com/caddyserver/caddy/releases/tag/v2.11.1),
[TLS automation options](https://caddyserver.com/docs/caddyfile/options#tls-automate-names).
