# `stepaccordion` and `chipgrid` — behaviour

Every diamond has an id. The traceability table in [`docs/PLAN.md`](../PLAN.md) §5 maps each
branch (`A1-yes`, `A1-no`, …) to the test that covers it. A branch without a test is a defect.

## 1. StepAccordion — step phases

```mermaid
stateDiagram-v2
    [*] --> open: first step after Init
    [*] --> locked: every other step after Init
    locked --> open: Advance() on the previous step
    open --> done: Advance(summary)
    done --> open: Reopen(key)
    done --> locked: Reopen() of an EARLIER step
    open --> locked: Reopen() of an EARLIER step
```

## 2. StepAccordion — `Advance(summary)`

```mermaid
flowchart TD
    A[Advance summary] --> A1{A1: is there an open step?}
    A1 -- no --> N[no-op]
    A1 -- yes --> D[mark it done, set its summary]
    D --> A2{A2: is it the last step?}
    A2 -- yes --> END[nothing open; Open returns ""]
    A2 -- no --> NX[open the next step]
```

## 3. StepAccordion — `Reopen(key)`

```mermaid
flowchart TD
    A[Reopen key] --> B1{B1: does a step with key exist?}
    B1 -- no --> N1[no-op]
    B1 -- yes --> B2{B2: is it done?}
    B2 -- no --> N2[no-op]
    B2 -- yes --> O[open it]
    O --> L[every later step: locked, summary cleared;<br>an open later step closes]
    L --> B3{B3: OnReopen set?}
    B3 -- yes --> CB[OnReopen key]
    B3 -- no --> END[done]
```

## 4. StepAccordion — render

```mermaid
flowchart TD
    A[Render] --> R1{R1: zero steps?}
    R1 -- yes --> E[empty root]
    R1 -- no --> S[one section per step]
    S --> R2{R2: Init not called yet?}
    R2 -- yes --> I[Init first] --> S2[bind states]
    R2 -- no --> S2
```

## 5. ChipGrid — pick and keyboard

```mermaid
flowchart TD
    A[click or Enter/Space on a chip] --> P1{P1: chip disabled?}
    P1 -- yes --> N[ignored]
    P1 -- no --> SEL[Selected = value]
    SEL --> P2{P2: OnPick set?}
    P2 -- yes --> CB[OnPick value]
    P2 -- no --> END[done]

    K[arrow / Home / End] --> K1{K1: is there an enabled chip<br>in that direction?}
    K1 -- no --> STAY[focus stays]
    K1 -- yes --> MOVE[focus moves, skipping disabled]
```

## 6. ChipGrid — `SetChips(chips)`

```mermaid
flowchart TD
    A[SetChips] --> Z1{Z1: zero chips?}
    Z1 -- yes --> EMPTY[render the Empty text<br>default 'No options available']
    Z1 -- no --> CH[render chips]
    CH --> Z2{Z2: is Selected among the<br>ENABLED chips?}
    Z2 -- yes --> KEEP[keep Selected]
    Z2 -- no --> CLR[Selected = empty]
    CH --> Z3{Z3: tab stop}
    Z3 -- a chip is selected --> T1[that chip tabindex 0]
    Z3 -- none selected --> T2[first enabled chip tabindex 0]
```
