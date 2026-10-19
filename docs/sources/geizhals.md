# Geizhals source notes

Findings from the spike on branch `spike/geizhals` (October 2026).

## Access rules

- `robots.txt` allows product pages (`/<slug>-a<ID>.html`) for `User-agent: *`.
  Category paths, `/redir/` and many query parameters (including `phist=`)
  are disallowed. There is no `Crawl-delay` for `*`.
- Rules for DealStalker:
  - Fetch only bare product URLs, never with query parameters.
  - Never fetch the price history page (`?phist=`).
  - Identify with an honest User-Agent that links to this repository.
  - Keep the request rate low; deduplicate requests across users.
- The Geizhals terms (AGB) forbid merchants from spidering the site. No
  explicit terms for regular visitors were found. Users of DealStalker are
  responsible for complying with the terms of the sites they track.

## Price extraction

- Prices are rendered server-side; no JavaScript is needed.
- The lowest listed price is available as
  `<meta property="og:price:amount" content="...">`.
- Observed formats: `"479"`, `"479.99"`, `"478.99"` (dot as decimal
  separator, zero or two decimals). Parse strictly into `int64` cents and
  reject anything else.
- `og:price:currency` is `EUR` on geizhals.de.
- The `<head>` is minified and attributes mix single and double quotes:
  use an HTML parser, not regular expressions.
- Pages without offers may not contain a price; this must be handled as
  "no price available", not as a crash.

## Product identity

- The `a<ID>` at the end of the URL identifies the exact product and is
  used as its unique key.
- The `canonical` link often points to a variant overview (`v<ID>`).
  Do not use it, or a different color or size may be tracked.

## Freshness

- Geizhals itself refreshes prices periodically (the page shows the time of
  the last update). Checking more often than about once per hour brings no
  benefit.