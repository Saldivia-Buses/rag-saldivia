# Reconnaissance — Tokens de claude.ai

**Fecha:** 2026-03-29
**Fuente:** Playwright getComputedStyle en https://claude.ai/login

## Body

- Background: `rgb(250, 249, 245)` → `#faf9f5`
- Foreground: `rgb(20, 20, 19)` → `#141413`
- Font: "Anthropic Sans" (propietaria, usamos Instrument Sans)

## Escala de grises (HSL, hue 45-60, warm)

| Token | HSL | Hex aprox | Uso |
|---|---|---|---|
| gray-10 | 60 14% 99% | #fdfdfa | — |
| gray-20 | 60 14% 97% | #f8f7f3 | — |
| gray-30 | 60 10% 96% | #f5f5f2 | — |
| gray-50 | 45 12% 93% | #efede8 | surfaces |
| gray-70 | 50 12% 91% | #eae7e0 | surface-2 |
| gray-80 | 50 11% 89% | #e5e2db | — |
| gray-100 | 53 12% 87% | #e0ddd5 | borders |
| gray-150 | 55 11% 80% | #d0cdc3 | border-strong |
| gray-300 | 55 6% 63% | #a4a19b | — |
| gray-500 | 40 3% 42% | #6e6c69 | fg-subtle |
| gray-700 | 60 3% 21% | #363633 | — |
| gray-800 | 60 2% 12% | #201f1e | dark bg |

## Bordes (computed)

- Input border: `1px solid rgba(31, 30, 29, 0.15)` → ~#e4e2de
- Button border (secondary): `1px solid rgba(31, 30, 29, 0.3)` → ~#c8c6c2

## Botones

- **CTA (Try Claude):** bg negro, color blanco, border-radius 9.6px, font-weight 500
- **Secondary (Contact sales):** bg transparente, color #141413, border 1px rgba(31,30,29,0.3), radius 8px
- **Nav text:** color #3d3d3a, weight 400

## Input

- bg: white (#ffffff)
- border: 1px solid rgba(31, 30, 29, 0.15)
- border-radius: 9.6px
- font-size: 16px

## Brand

- `--_brand-clay`: HSL 14.8 63.1% 59.6% (naranja Anthropic, NO usamos)

## Observaciones

1. Claude usa grises MUY warm (hue 45-60) — similar a nuestro approach actual
2. La saturación es baja (3-14%) — sutil, no obvio
3. Los bordes son semi-transparentes con negro, no colores sólidos
4. Los CTA son NEGROS, no azules — nosotros usamos azure blue como acento
5. La font es propietaria (Anthropic Sans), mantenemos Instrument Sans
6. Border-radius ~9.6px (0.6rem) — nosotros usamos 0.5rem (8px)
