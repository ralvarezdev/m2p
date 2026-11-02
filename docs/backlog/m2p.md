# m2p backlog

Queue for this repo, created 2026-11-01 from the [reuse audit](../../../../../libraries/go/identity/docs/reviews/2026-11-01-reuse-audit.md).
It consumes no house kit today; the rows are optional adoptions, not duplication found.

## Someday

- [ ] CLI presentation on `termkit` (`internal/cli`, 6 files): the theme, the table renderer and
  the capability detection are what the other CLIs already share; only take it if the output grows
  past plain lines  #m2p  P3
- [ ] Process lifecycle on `svckit` if `convert` grows a long-running mode: today it is a batch job,
  so `svckit` buys nothing  #m2p  P3
- [ ] If the converter ever persists state, use `sqlitekit` (pure-Go driver, WAL and pragmas on
  every connection, embedded goose migrations) instead of opening `database/sql` by hand  #m2p  P3
