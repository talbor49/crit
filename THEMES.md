# Contributing a theme

Crit ships with **System / Light / Dark** plus a set of community palettes
(Catppuccin Mocha & Latte, Dracula, Nord, Gruvbox Dark, Rosé Pine). Adding
another one is two files and about forty lines of CSS — no build step, no
JavaScript beyond one registry entry.

Pick your palette from **Settings → Display → Palette**.

---

## How theming works

`web/theme.css` defines every colour as a CSS custom property. There are two
families:

| Family | Applies to | Examples |
| --- | --- | --- |
| `--crit-*` (brand) | header, sidebar, chips, buttons, modals | `--crit-bg-page`, `--crit-brand`, `--crit-pill-bg` |
| `--crit-editor-*` | the review surface: diffs, code, comment cards | `--crit-editor-bg`, `--crit-editor-fg`, `--crit-editor-code-bg` |

Themes are selected by an attribute on `<html>`:

- no `data-theme` → follow the OS (`prefers-color-scheme`)
- `data-theme="light"` / `data-theme="dark"` → the built-ins
- `data-theme="<your-id>"` → your palette

The **cascade is what keeps a theme short.** `:root` holds the complete dark
token set, and `[data-theme="light"]` holds the complete light set. Your block
comes later in the file, so it only needs the tokens that make your palette
recognisable — roughly forty of the ~170. Everything you leave out falls
through to the base.

Syntax highlighting works the same way: you set nine `--crit-syn-*` tokens and
one generic `html.crit-theme-custom .hljs-*` block at the bottom of `theme.css`
maps them onto highlight.js classes. **You never write highlight.js rules.**

---

## Adding a palette

### 1. Register it

`web/crit-shared.js`, in the `THEMES` array:

```js
{ id: 'tokyo-night-storm', label: 'Tokyo Night Storm', dark: true },
```

- `id` — kebab-case; it becomes the `data-theme` value and the cookie value.
- `label` — what the Settings dropdown shows.
- `dark` — `true` for a dark palette, `false` for a light one. It drives the
  mermaid diagram theme and the light-mode contrast fixes; get it wrong and
  diagrams render against the wrong background.

### 2. Light palettes only: inherit the light base

A `[data-theme="…"]` block does **not** match
`@media (prefers-color-scheme: light)`, so a light palette that skips this step
inherits the *dark* base and looks broken in every corner you didn't override.

Add your id to the selector list at the top of the explicit-light block in
`web/theme.css` (keep `[data-theme="light"] {` as the last line of the
selector — a script parses for it):

```css
[data-theme="catppuccin-latte"],
[data-theme="your-light-theme"],
[data-theme="light"] {
```

Dark palettes need nothing here: `:root` is already the dark base.

### 3. Write the palette block

Add it to the community section of `web/theme.css`, after the existing ones.
Copy this template and fill in every value:

```css
/* ---- Your Theme (dark) ---- */
[data-theme="your-theme"] {
  /* Chrome surfaces — page background, cards, raised elements */
  --crit-bg-page:     #000000;
  --crit-bg-card:     #000000;
  --crit-bg-elevated: #000000;

  /* Chrome text */
  --crit-fg-primary:   #000000;
  --crit-fg-secondary: #000000;
  --crit-fg-muted:     #000000;

  /* Accent. *-subtle and *-bg are the same hue at 10% / 15% alpha. */
  --crit-brand:         #000000;
  --crit-brand-hover:   #000000;
  --crit-brand-subtle:  rgba(0, 0, 0, 0.10);
  --crit-brand-bg:      rgba(0, 0, 0, 0.15);
  --crit-line-focus-bg: rgba(0, 0, 0, 0.10);
  --crit-brand-cta:       #000000;
  --crit-brand-cta-hover: #000000;

  /* Header chips */
  --crit-pill-bg:       #000000;
  --crit-pill-bg-hover: #000000;
  --crit-pill-fg:       #000000;
  --crit-pill-fg-muted: #000000;
  --crit-pill-border:   #000000;
  --crit-rail-fg:       #000000;
  --crit-rail-arrow:    #000000;

  /* Review surface */
  --crit-editor-bg:          #000000;
  --crit-editor-bg-card:     #000000;
  --crit-editor-bg-elevated: #000000;
  --crit-editor-bg-hover:    #000000;
  --crit-editor-bg-gutter:   #000000;
  --crit-editor-code-bg:     #000000;
  --crit-editor-comment-bg:  #000000;
  --crit-editor-fg:           #000000;
  --crit-editor-fg-secondary: #000000;
  --crit-editor-fg-muted:     #000000;
  --crit-editor-scrollbar-bg:    #000000;
  --crit-editor-scrollbar-thumb: #000000;

  /* Semantic. These also colour comment severity badges:
     red = Critical, yellow = Important, green = Suggestion. */
  --crit-green:  #000000;
  --crit-red:    #000000;
  --crit-orange: #000000;
  --crit-yellow: #000000;
  --crit-purple: #000000;
  --crit-visual-accent: #000000;

  /* Borders */
  --crit-border:         #000000;
  --crit-border-strong:  #000000;
  --crit-border-comment: #000000;

  /* Syntax highlighting */
  --crit-syn-comment:     #000000;
  --crit-syn-keyword:     #000000;
  --crit-syn-string:      #000000;
  --crit-syn-number:      #000000;
  --crit-syn-function:    #000000;
  --crit-syn-type:        #000000;
  --crit-syn-tag:         #000000;
  --crit-syn-builtin:     #000000;
  --crit-syn-punctuation: #000000;
}
```

That's the whole theme. There is no step 4.

---

## Token reference

### Surfaces, darkest to lightest (dark palettes; invert for light)

```
--crit-bg-page        the page behind everything
--crit-bg-card        panels, the file tree, modals
--crit-bg-elevated    raised chrome: menus, chips, hover states
```

`--crit-editor-*` mirrors these for the review surface. Keeping
`--crit-editor-bg` a touch different from `--crit-bg-page` is what makes the
diff read as a distinct surface — most palettes use their "base" for the editor
and their "mantle/crust" for the page.

### The nine syntax tokens

| Token | highlight.js groups |
| --- | --- |
| `--crit-syn-comment` | `comment`, `meta` |
| `--crit-syn-keyword` | `keyword`, `operator`, `attr`, `name` |
| `--crit-syn-string` | `string`, `quote`, `bullet` |
| `--crit-syn-number` | `number`, `literal`, `variable`, `params`, `link` |
| `--crit-syn-function` | `title`, `function_`, `class_`, `property`, `section` |
| `--crit-syn-type` | `type`, `selector-tag` |
| `--crit-syn-tag` | `tag`, `doctag`, `regexp`, `selector-id`, `selector-class` |
| `--crit-syn-builtin` | `built_in`, `attribute`, `symbol` |
| `--crit-syn-punctuation` | `punctuation`, and the default code colour |

If the upstream palette publishes an editor theme, lift these straight from it
rather than guessing.

### What you can also override

Anything in `:root` is fair game — diff backgrounds
(`--crit-diff-add-bg`, `--crit-diff-del-line-bg`, …), author swatches
(`--author-0-bg` … `--author-5-fg`), live-mode markers. The defaults are
translucent overlays that work on most backgrounds, so start without them and
add only what looks wrong.

Do **not** override `--crit-r-*` (radii), `--crit-dur-*` / `--crit-ease*`
(motion), or the font stacks. Those are product decisions, not palette ones;
fonts have their own setting.

---

## Checking your work

```bash
./scripts/check-css-vars.sh    # no undefined, dead, or missing tokens
npx stylelint web/*.css        # no hardcoded colours outside theme.css
go build -o crit ./cmd/crit && ./crit
```

Then, with crit open, walk this list in your palette:

- [ ] File tree, header chips, and the settings overlay are all legible
- [ ] A code diff: additions, deletions, and **word-level** highlights inside
      them stay readable — this is where light palettes usually fail
- [ ] Comment cards, including a `[Critical]` / `[Important]` / `[Suggestion]`
      severity badge on each
- [ ] A resolved comment and a drifted comment
- [ ] Markdown document view: headings, tables, blockquotes, inline code
- [ ] Text contrast is at least **4.5:1** against its background (WCAG AA), and
      **3:1** for borders and icons

The last one is not negotiable — a palette that looks great and can't be read
in daylight helps nobody.

---

## Submitting

One theme per pull request, titled `feat(theme): <name>`. Include a screenshot
of a code diff and one of a markdown document. If the palette is an established
one, link its upstream spec so the hex values can be checked against the source.

Palettes are accepted on the same terms as the rest of crit: they should be
something you'd actually spend a day reviewing code in.
