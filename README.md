# Community Nongrams

Community Nongrams is a Go + Ebitengine nonogram game with offline puzzles, a pixel-art editor, local drafts, packs, creator profiles, and an optional Supabase-backed community gallery.

The game can run completely offline. Supabase is only needed for accounts, publishing, gallery likes, play counts, chat, creator profiles, and main-game review submissions.

<img width="1183" height="1376" alt="Community Nongrams gameplay" src="https://github.com/user-attachments/assets/2315be46-67d8-43d6-9ac9-fa36fb654cb8" />

## Run

Desktop:

```sh
go run ./cmd/game
```

Web dev loop:

```sh
PORT=8000 ./scripts/dev-web.sh
```

That writes `static/config.js`, regenerates levels, builds `static/game.wasm`, serves `static/`, and rebuilds when Go, level, asset, or HTML files change. If `watchexec` is not installed, the script falls back to polling.

Vercel build:

```sh
scripts/build-vercel.sh
```

Vercel is configured by `vercel.json` to run that command and deploy `static/`.

Controls:

- Click or touch the right-side trigger to switch between fill and X-mark tools.
- Drag across cells to keep applying the selected tool.
- `F` selects fill, `X` or `M` selects X-mark, `Z` undoes, and `R` resets.

## Features

- Offline levels generated from image sheets in `levels/`.
- Community art editor with before/final-art layers, drawing tools, import, export, undo, title editing, and size changes.
- Local browser drafts for guest creators.
- Pack creation from published art.
- Community gallery with art and packs, sort by new, most played, or top rated.
- Likes, play counts, and per-art/per-pack chat when Supabase is configured.
- Creator profiles with avatar art, bio, display name, social icons, favorite palette, favorite color, and promoted work.
- Optional main-game review flow for published art.

## Add Offline Levels

Create a folder in `levels/` named with the level number and title:

```text
levels/007-flower/
  art.png
```

The image can be named `art.png`, `art.webp`, `sheet.png`, or `sheet.webp`. If there is only one PNG or WebP in the folder, that file is used automatically.

The image must be a two-panel spritesheet: the left tile is the before/line-art puzzle source, and the right tile is the colored reveal. A 10x10 level should be a 20x10 image, and a 15x15 level should be a 30x15 image. The generator infers the tile size from the image height.

The older flat-file format still works:

```text
levels/L3-Flower_16.png
```

In that format, the suffix is the tile size. A `_16` file should be 32x16. If the suffix is wrong but the image is still a two-panel sheet, the generator uses the image height as the tile size.

For opaque black-and-white line art, the generator treats the dominant white-ish or black-ish first-panel color as the empty background.

Generate puzzle JSON from every sheet in `levels/`:

```sh
go run ./cmd/genlevels
```

This writes self-contained `puzzle.json` files under `assets/puzzles/` and `internal/assets/embedded/assets/puzzles/`. No split skeleton/reveal images are generated.

## Community Editor

Open **Community > Create** to draw new art. The editor treats the first frame/layer as the before puzzle source and the second frame/layer as the final reveal art. The before layer determines the nonogram solution; the final-art layer keeps its colors.

Guest drawings and packs are saved in browser local storage. **My Art** keeps multiple editable drafts. Publishing asks for name, description, tags, and an optional main-game review request. Main-game review requires a rights confirmation and does not guarantee approval.

Published art can be bundled into packs. Packs can include local art directly, and published packs can be edited later from the Published tab.

### Import From Aseprite

For the most reliable multi-level import, export the sprite sheet as PNG and enable Aseprite's JSON data export. Name paired frames with the same base name:

```text
flower_before
flower_after
lion_before
lion_after
```

Select the PNG and JSON together from **Community > My Library > + > Import Sprite Sheet**. Frames must be square and 8, 10, 15, 20, or 32 pixels. The importer turns every matching `_before` / `_after` pair into a separate draft.

For a regular PNG without JSON, arrange any number of pairs horizontally (`Before, After`) or vertically (`Before` above `After`). The image dimensions must be exact multiples of the selected tile size. A single 1x2 pair works through the same importer.

## Community Backend

Copy the example environment file:

```sh
cp .env.example .env
```

Set these values for local web community features:

```sh
SUPABASE_URL=https://YOUR_PROJECT.supabase.co
SUPABASE_ANON_KEY=YOUR_PUBLIC_ANON_KEY
```

`SUPABASE_URL` should be the project URL. If you paste a Data API URL ending in `/rest/v1`, `scripts/write-web-config.sh` normalizes it. Do not use a service-role key in the browser config.

Setup steps:

1. Create a Supabase project.
2. For a fresh project, run `supabase/schema.sql`, then run migrations `029_reliable_content_chat.sql` and `030_database_hardening.sql`. Existing projects should apply every numbered migration they have not yet applied. Migration 030 is the final authority for database permissions and limits.
3. Run `scripts/write-web-config.sh` before serving `static/`; `scripts/dev-web.sh` does this automatically.
4. Add the game URL to Supabase Auth redirect URLs and enable email magic links.
5. Optional: configure Google OAuth using the values in `.env.example`.
6. Deploy `supabase/functions/notify-official-submission` and set `WEBHOOK_SECRET`, `RESEND_API_KEY`, `REVIEW_EMAIL`, `REVIEW_FROM_EMAIL`, and `ADMIN_REVIEW_URL`.
7. Create a Supabase Database Webhook for inserts on `official_submissions`, targeting that Edge Function with the matching `x-webhook-secret` header.
8. Set reviewer profiles to `moderator` or `admin` with a trusted SQL/admin operation. Review submissions at `/admin.html` after signing in through the game.

Published level versions are immutable. Packs reference exact versions, user data is protected by row-level security, and review emails contain an admin link instead of artwork attachments.

## Deploy To Vercel

The repo is ready to deploy as a static Vercel project.

Vercel settings:

- Framework Preset: Other
- Build Command: `scripts/build-vercel.sh`
- Output Directory: `static`
- Install Command: empty

Required Vercel environment variables for community cloud:

```sh
SUPABASE_URL=https://YOUR_PROJECT.supabase.co
SUPABASE_ANON_KEY=YOUR_PUBLIC_ANON_KEY
```

If those values are omitted, the game still deploys and offline levels still work, but accounts, publishing, gallery, chat, and profiles stay disabled.

After deployment, add the Vercel production URL and any preview URLs you use to Supabase Auth redirect URLs.

## Supabase Keepalive

Free-tier Supabase projects can pause after 7 days of inactivity. This repo includes `.github/workflows/supabase-keepalive.yml`, which runs twice a week and makes a small read request to the Supabase Data API.

Add these GitHub Actions repository secrets:

```sh
SUPABASE_URL=https://YOUR_PROJECT.supabase.co
SUPABASE_ANON_KEY=YOUR_PUBLIC_ANON_KEY
```

You can also run the workflow manually from the GitHub Actions tab after unpausing a project.

## Checks

Local checks:

```sh
gofmt -w ./cmd ./internal
go test ./internal/community ./internal/nonogram ./internal/pixelpuzzle ./cmd/genlevels
go run github.com/go-critic/go-critic/cmd/gocritic@latest check -enable='#diagnostic' -disable=commentedOutCode ./...
GOOS=js GOARCH=wasm go build -buildvcs=false -o static/game.wasm ./cmd/game
```

The full native `go test ./...` path can hit Ebitengine display dependencies on machines without X11/graphics setup. The focused package test list above avoids that while still covering the nonogram, pixel puzzle, community model, and level generator code.

## Web And Mobile

Manual WebAssembly build:

```sh
GOOS=js GOARCH=wasm go build -buildvcs=false -o static/game.wasm ./cmd/game
```

Android and iOS can be added once the MVP input, layout, and asset loading choices are stable.

<img width="813" height="1174" alt="Community Nongrams editor" src="https://github.com/user-attachments/assets/5a22b50c-164d-4e11-b366-419a6ed7ea4b" />
