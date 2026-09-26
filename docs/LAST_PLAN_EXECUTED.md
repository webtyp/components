---
PLAN: "refactor(components): consume the typed event and family contracts; pay the calendarslider debt"
EXECUTOR: unassigned
REVIEWER: none
---

> **Phase C** of
> [`webtyp/docs/TYPED_EVENTS_AND_SURFACES_MASTER_PLAN.md`](https://github.com/webtyp/webtyp/blob/main/docs/TYPED_EVENTS_AND_SURFACES_MASTER_PLAN.md).
> **Blocked on two gates.** Bump `go.mod` to the published tags of
> [`webtyp.com/dom`](../../dom/docs/PLAN.md) (Phase A) and
> [`webtyp.com/widget`](../../widget/docs/PLAN.md) (Phase B) **before** stage 1.
> Also: `components/go.mod` currently carries a local
> `replace webtyp.com/widget => ../widget` used during development —
> **delete it in stage 0.**

# Plan — assembly only, no new API

No exported symbol is added in this repo, so the api-design gate does not apply
here. Every gate answer lives in the two phase plans above. What this phase does
is stop hand-composing what those libraries now own, and pay one debt that
Phase A finally makes payable.

Files touched: `go.mod`, `conformance_test.go`, `calendarslider/`,
`targethour/css.go`, `usermenu/css.go`, plus the mechanical `On(` migration
across every package.

## 1. The debt this phase pays

`calendarslider` currently ships `role="grid"` on its strip
(`calendarslider.go`, the `strip` element) and `role="gridcell"` on every day.
The WAI-ARIA grid pattern that role names **requires** arrow-key navigation.
The component does not implement it, and could not: `dom.Event` had no way to
read a key press.

Measured in the running demo on 2026-09-10: focus does not move on
`ArrowRight`, `ArrowDown` or `Home`; all day buttons carry `tabIndex: 0`. What
looks like arrow navigation is the browser scrolling the `overflow-x: auto`
strip natively.

This is the escalation `CONSTRUCTION_HARNESS.md` now records: one untyped
parameter in `dom` became an accessibility promise a leaf component cannot keep.
Phase A removes the cause; this phase removes the symptom. **`role="grid"` is
not allowed to survive this plan without working arrow keys** — see the master
plan's closing conditions.

## 2. Stages

### Stage 0 — unblock

- Delete `replace webtyp.com/widget => ../widget` from `go.mod`.
- Bump `webtyp.com/dom` and `webtyp.com/widget` to the tags Phases A and B
  published. `gotest` must be green before anything else starts.

### Stage 1 — consume the typed events

Mechanical, whole repo: `On("click", f)` → `OnClick(f)`, `On("change", f)` →
`OnChange(f)`, `On("input", f)` → `OnInput(f)`, and the same for `blur`,
`toggle`, `submit`. 47 non-test call sites; the compiler finds every one, since
`On` no longer exists.

Done when `grep -rn 'On("' --include='*.go' .` returns nothing.

### Stage 2 — consume the family contract

Six blocks rely on `Interactive(s)` implicitly painting `s` and must now declare
their resting surface explicitly. The compiler does not catch this — the parts
simply lose their background — so verify each visually, not only by test.

Two blocks change meaning and are the reason Phase B exists:

- `targethour/css.go` `PartFree` — keeps `As(Subtle)` and `Interactive(Page)`,
  and now finally gets what its own comment already claims: *"the row keeps its
  Interactive(Page) surface over the whole box"*.
- `usermenu/css.go` `PartTrigger` — `Button(Subtle)` keeps the ghost resting
  look and gains a legible hover. Verify the measured ratio: it must be ≥ 4.5:1,
  not the 1.42:1 it ships today.

### Stage 3 — give the cap rule back to its owner

Delete `TestNoHandRolledIconCaps` from `conformance_test.go`, including its
`targetdate/PartLead` allowlist entry. Phase B moved the rule into `Validate()`,
where it covers every consumer instead of this repo. Two copies of a rule drift.

`targetdate`'s `PartLead` must still validate after the move — it is a square
control-height block holding **text**, not a glyph cap. If Phase B's diagnostic
rejects it, that is a defect in Phase B's rule, not a reason to re-add an
allowlist here: report it upstream.

### Stage 4 — calendarslider keyboard, on the real contract

Implement the WAI-ARIA grid pattern the role already promises, using
`dom.OnKeyDown` and `dom.Key`:

- **Roving tabindex.** Exactly one day per month card carries `tabindex="0"` —
  the selected day, else today, else the first selectable day. Every other day
  carries `tabindex="-1"`. Today all of them are `0`.
- **Arrow keys** on the month card: `KeyArrowLeft`/`KeyArrowRight` move ±1 day,
  `KeyArrowUp`/`KeyArrowDown` ±7, `KeyHome`/`KeyEnd` to the ends of the week.
  Focus moves with `Ref().Focus()`, which already exists and already suppresses
  the browser's own scrolling (`element_wasm.go`).
- Movement stays inside the month card. Crossing a month boundary is **not** in
  this plan: it needs a decision about whether it should slide the strip, and an
  undecided thing does not get a half implementation.

If stage 4 cannot be completed, the fallback is **not** to leave it: downgrade
the strip to `role="group"` and the cells to plain list items in the same
change, so the markup stops claiming what it does not do.

### Stage 5 — verify in the browser, not only in tests

With the MCP browser tools, on the `calendarslider` demo and on the reservation
module in `app-demo`:

1. Focus the first selectable day; press `ArrowRight`, `ArrowDown`, `Home`.
   `document.activeElement` must change each time. This is the exact assertion
   that fails today.
2. `Tab` must reach the calendar and leave it in **one** stop per month card,
   not 31.
3. `browser_audit_mobile` on `.calendarslider`: no regression against the
   current report.
4. Hover `usermenu`'s trigger and `targethour`'s free row: text legible in both
   themes.

## 3. Zero technical debt — closing check

- `grep -rn 'On("' --include='*.go' .` → nothing.
- `grep -rn 'replace ' go.mod` → nothing.
- `grep -rn 'TestNoHandRolledIconCaps' .` → nothing.
- No `role="grid"` without working arrow keys.
- No component composing a cap from `MediaBox(AspectSquare)` + `ControlBox()`.
- `grep -rn "TODO\|FIXME\|Deprecated" --include='*.go' .` — every hit predates
  this change.
