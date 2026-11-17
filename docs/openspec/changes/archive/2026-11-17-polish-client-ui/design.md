# Design

## Context

The client is SvelteKit with Tailwind v4 and flowbite-svelte. Pages set colours inline (`text-white`, `text-gray-300`), which only works on a dark background, while the body background is a light stone in both themes. Dark mode is toggled by flowbite's `DarkMode` (a `.dark` class on `<html>`), but only flowbite components react to it. A global `form` rule in `app.css` boxed every form, including inline and dialog forms. Every page renders its own `NavBar` and `Footer`; there is no shared content container. See proposal.md for the list of visible problems.

## Goals / Non-Goals

**Goals:**

- One place that decides colours for both themes, so a page cannot be readable in only one of them
- Shared building blocks for the repeated parts: content container, raised panel, action buttons
- Keep every label, heading, button name and test id the Playwright tests select on

**Non-Goals:**

- Replacing flowbite or redesigning the brand colours (orange bar, stone neutrals)
- The full-screen views (`/views/devices`, `/devices/pair`), which draw their own black screen
- Behaviour changes of any kind

## Decisions

**Semantic colour tokens in `app.css`.** `--color-page`, `surface`, `ink`, `muted`, `line`, `accent` and `placeholder` are defined in `@theme` (light) and overridden in a `.dark` block. Tailwind v4 emits utilities as `var(--color-…)`, so `bg-page` or `text-muted` switch theme without a `dark:` variant. Alternative: `dark:` variants on every element, which is what produced the half-working dark mode. Values were picked to reach WCAG AA: ink 16:1, muted at least 7:1, accent at least 4.6:1 on page and surface in both themes. Where a flowbite component sets its own `dark:text-*`, the page also passes `dark:text-ink` / `dark:text-muted`, because the `dark:` class otherwise wins.

**Tailwind `gray` remapped to stone.** flowbite paints inputs, tables, cards and dialogs with `gray-*`. Overriding `--color-gray-*` with the stone values makes every flowbite component match the page without theming each component. Alternative: flowbite theme overrides per component, which is far more code for the same result.

**Placeholder rule outside any layer.** flowbite sets no light-mode placeholder colour, so the browser default looked like a filled-in value. An unlayered `::placeholder` rule beats Tailwind's layered utilities without `!important`.

**`ActionButton` wraps flowbite `Button`.** An `intent` prop (`create`, `delete`, `neutral`) picks the hover colour (green, red, darker stone) on a dark default; disabled is grey with `opacity-100`, so it reads as inert rather than faded. Hover uses `enabled:hover:` so a disabled button does not light up. Alternative: global CSS classes, but tailwind-merge in flowbite only resolves conflicts for classes passed through `class`.

**`Page` container and a full-height root layout.** `+layout.svelte` is a `min-h-screen` flex column; `Page` is a centred `<main>` with gutters, a `width` (`narrow`, `default`, `wide`) and `grow`, and the footer keeps `mt-auto`. Alternative: moving `NavBar` and `Footer` into the root layout, which would also put them on the full-screen views. The global `form` rule became an opt-in `.panel` class.

**Nav highlights the current page.** `NavUl` gets `activeUrl` from `page.url`, with the trailing slash stripped because `trailingSlash = 'always'` and the links have none. Link text is dark stone on the orange bar, as white reaches only 3.7:1.

**Size check draws into its own canvas.** The demo passes a bound `<canvas>` inside a fixed, full-viewport container to Pixi and resizes to that container. It used to append the canvas to `<body>` below the page and size it to `window`, whose `innerWidth` includes a scrollbar: that was the too-high, too-wide result. The page shows the canvas and viewport size, and destroys the Pixi app when you leave.

## Risks / Trade-offs

- [Remapping `gray` changes every `gray-*` use, including future ones] → Intended; stone is the only neutral in the palette. A cool grey can still be added under another name.
- [Placeholder contrast is about 2.5:1] → Placeholders are hints, not content, and WCAG does not require AA for them; a lighter placeholder was the point of the task.
- [Pages built later can still hardcode colours] → The tokens, `Page`, `.panel` and `ActionButton` are the documented way; reviewers can check for `text-white` / `text-gray-*` on page chrome.
