# Design

## Context

There is no mature Go Matter library, so this means writing a controller: commissioning and fabric management.

## Decisions

- Native Go controller rather than delegating to another hub

## Risks / Trade-offs

- Large scope and protocol complexity → land commissioning and on/off/colour first, behind its own config flag
