# Agent Note: Anthropic Messages browser preflight headers

Status: implemented

## Problem

The detector's browser-direct Messages request sends `x-api-key`, `anthropic-version`, and `anthropic-dangerous-direct-browser-access`; its beta path also sends `anthropic-beta`. The CORS middleware allows the trusted Origin and answers OPTIONS with 204, but omits those three Anthropic headers. The browser therefore prevents POST and reports Failed to fetch even though the two OpenAI protocols work.

## Decision

Add only the three Anthropic headers to the backend CORS allowlist. Preserve the configured exact Origin policy, authentication, private Tailscale access, and existing SDK headers. Test header membership case-insensitively instead of treating OPTIONS 204 as sufficient. An untrusted Origin must still receive 403 without CORS allow headers.

Build the deployment binary from the live production source commit with only this middleware change. Reuse production-verified embedded frontend assets and derive a local image from the retained original runtime image. Recreate only the application service; preserve configuration and all other containers.

## Alternatives considered

- Removing Anthropic headers in the detector hides the server compatibility issue and alters required request semantics.
- Globally injecting permissive CORS in the reverse proxy duplicates policy and broadens the access boundary.
- Routing private API calls through public Pages Functions breaks the intended browser-direct private-network design.
- Deploying the local development tree risks unrelated changes; use the production source base instead.

## Consequences

Allowed Origins can send the normal and beta Messages request shapes. This does not authorize additional Origins or bypass API authentication. Browser acceptance deliberately uses an invalid placeholder key and an empty body; a readable 401 proves browser transport/CORS, not successful model generation.

Deployment packaging must verify execution as the actual non-root runtime user. Root-only `-version` misses a 0700 binary copied by archive extraction under umask 077. Explicit 0755 mode and an isolated entrypoint startup check prevent recurrence. Preserve the first failed attempt and its successful rollback separately from the subsequent release result.

## Testing

The regression test first fails for precisely the three missing headers, then the entire middleware package passes with the whitelist change. Real HTTPS preflight checks cover Messages, Responses, Chat Completions and an untrusted Origin. Real BrowserOS checks exercise normal Messages, beta Messages and both OpenAI shapes without accessing a real key. Deployment and acceptance receipts live in the detector's local direct-repair evidence directory; they, not this design note, establish current production state.

## Existing note audit

No existing decision-note tree, ADR directory, or decision directory was present when creating this note; no overlapping notes were found or removed.

## Source integration and upgrade boundary (2026-10-22)

The earlier deployment used a local derived image, not an upstream release. Updating to an image without this source patch can therefore restore the same missing-header failure. The owner now requests integrating the existing middleware fix and its regression test into the local `dev` branch. Preserve the exact-Origin policy and existing OpenAI headers; include this note so future maintainers know why the three additional headers are necessary.

This source integration does not publish a container, push Git changes, or repair an already-running deployment. Build subsequent images from a revision containing the fix and run the protocol preflight checks after upgrades. Revalidate the regression using a Go source overlay with the unpatched HEAD version, avoiding any rollback of the live working tree, then run the complete middleware test package with the patched source.
