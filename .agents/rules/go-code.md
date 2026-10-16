# Go code

Refines `~/.claude/rule-details/go/wrapping-errors.md` and `~/.claude/rule-details/coding-style.md`.

- **GO1** `client.gen.go` and the `.gen.go` files in `ws/` and `perpsws/` are generated. Never edit them by hand. Change `config.yaml`, `overlay.yaml`, the package `config.yaml` files or the spec patch in `scripts/`, then run `make generate` and commit the result.
- **GO2** Spec fixes live in `scripts/` or `overlay.yaml`, applied before generation. `scripts/match-client-spec.sh` turns Kalshi's server-side AsyncAPI documents into client-side ones; `overlay.yaml` fixes the OpenAPI document. Each fix says in a comment which spec problem it solves.
- **GO3** The downloaded specs in `api/` are gitignored and never committed.
- **GO4** `auth/` is hand-written. Its errors wrap `ctxerrors/commerr` sentinels with `ctxerrors.Wrap` or `ctxerrors.Wrapf`. No `fmt.Errorf`, no `errors.New` outside a sentinel declaration. It never logs a key, a signature or a header value.
- **GO5** The exported API is what the generators emit plus `kalshiauth`. A change that renames or removes an exported name is a breaking release (see `releases.md`).
- **GO6** `make lint` passes with zero issues. Hand-written lines are at most 80 columns, the `.golangci.yml` limit.
