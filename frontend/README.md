# Frets frontend

React, TypeScript, Vite and Gravity UI. Modules are organized by responsibility;
`api` contains HTTP requests, `model` contains session state and hooks, and `ui`
contains components. Pages compose modules; shared code does not import modules.
Only the session module is implemented in this first increment.

## Development

Requires Node.js 22.12+ and pnpm 10.17.1.

```sh
pnpm install --frozen-lockfile
pnpm dev
```

The app runs on http://localhost:5173 and proxies `/api` to the backend at
http://127.0.0.1:6666. For logout and the development OAuth flow, set the backend
`PUBLIC_ORIGIN=http://localhost:5173` and register callback URLs on that same
origin, then recreate the backend containers. Production Compose serves both
frontend and API through Nginx, with the existing single `PUBLIC_ORIGIN`.

The login screen uses Yandex ID and Google. OAuth credentials are configured
only in the backend. An existing session shows the signed-in account and logout;
song library and playback follow in later increments. No local-login button is
exposed by the frontend.

```sh
pnpm typecheck
pnpm lint
pnpm format:check
pnpm test
pnpm build
```

The light theme maps the Figma palette to Gravity UI semantic tokens. The login
screen has no raster assets and makes no third-party image requests. Add future
Figma artwork as optimized local assets rather than temporary Figma URLs.
