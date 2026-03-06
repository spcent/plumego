#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

mkdir -p "$ROOT"/{monogram,toolkit,wordmark}/{svg,png} "$ROOT"/final/{svg,png}

variant_bg() {
  case "$1" in
    color) echo "$2" ;;
    bw) echo "#FFFFFF" ;;
    reverse) echo "$3" ;;
    *) echo "unknown variant: $1" >&2; exit 1 ;;
  esac
}

variant_fg() {
  case "$1" in
    color) echo "$2" ;;
    bw) echo "#111111" ;;
    reverse) echo "#FFFFFF" ;;
    *) echo "unknown variant: $1" >&2; exit 1 ;;
  esac
}

monogram_square() {
  local variant="$1"
  local bg fg accent
  bg="$(variant_bg "$variant" "#FFFFFF" "#0B1220")"
  fg="$(variant_fg "$variant" "#1F2937")"
  accent="$(variant_fg "$variant" "#00ADD8")"

  cat > "$ROOT/monogram/svg/monogram-square-${variant}.svg" <<SVG
<svg xmlns="http://www.w3.org/2000/svg" width="1024" height="1024" viewBox="0 0 1024 1024" role="img" aria-label="plumego monogram square ${variant}">
  <rect width="1024" height="1024" rx="168" fill="${bg}"/>
  <rect x="168" y="168" width="688" height="688" rx="140" fill="none" stroke="${accent}" stroke-width="34"/>
  <path d="M362 268v488" stroke="${fg}" stroke-width="68" stroke-linecap="round"/>
  <path d="M362 318h208c108 0 180 62 180 156s-72 156-180 156H362" fill="none" stroke="${fg}" stroke-width="68" stroke-linecap="round" stroke-linejoin="round"/>
  <circle cx="624" cy="473" r="20" fill="${bg}"/>
  <circle cx="672" cy="473" r="20" fill="${bg}"/>
  <circle cx="648" cy="432" r="20" fill="${bg}"/>
  <circle cx="648" cy="514" r="20" fill="${bg}"/>
  <circle cx="707" cy="432" r="20" fill="${bg}"/>
</svg>
SVG
}

monogram_horizontal() {
  local variant="$1"
  local bg fg accent
  bg="$(variant_bg "$variant" "#FFFFFF" "#0B1220")"
  fg="$(variant_fg "$variant" "#1F2937")"
  accent="$(variant_fg "$variant" "#00ADD8")"

  cat > "$ROOT/monogram/svg/monogram-horizontal-${variant}.svg" <<SVG
<svg xmlns="http://www.w3.org/2000/svg" width="1600" height="600" viewBox="0 0 1600 600" role="img" aria-label="plumego monogram horizontal ${variant}">
  <rect width="1600" height="600" rx="72" fill="${bg}"/>
  <rect x="92" y="92" width="416" height="416" rx="92" fill="none" stroke="${accent}" stroke-width="24"/>
  <path d="M196 150v300" stroke="${fg}" stroke-width="38" stroke-linecap="round"/>
  <path d="M196 182h132c72 0 120 38 120 96s-48 96-120 96H196" fill="none" stroke="${fg}" stroke-width="38" stroke-linecap="round" stroke-linejoin="round"/>
  <circle cx="368" cy="274" r="12" fill="${bg}"/>
  <circle cx="396" cy="274" r="12" fill="${bg}"/>
  <circle cx="382" cy="250" r="12" fill="${bg}"/>
  <circle cx="382" cy="298" r="12" fill="${bg}"/>
  <circle cx="416" cy="250" r="12" fill="${bg}"/>
  <text x="592" y="348" fill="${fg}" font-size="176" font-family="Avenir Next, SF Pro Display, Segoe UI, Arial, sans-serif" font-weight="700" letter-spacing="-3">plumego</text>
</svg>
SVG
}

toolkit_square() {
  local variant="$1"
  local bg fg accent muted
  bg="$(variant_bg "$variant" "#E5E7EB" "#111827")"
  fg="$(variant_fg "$variant" "#111827")"
  accent="$(variant_fg "$variant" "#0F766E")"
  muted="$(variant_fg "$variant" "#4B5563")"

  cat > "$ROOT/toolkit/svg/toolkit-square-${variant}.svg" <<SVG
<svg xmlns="http://www.w3.org/2000/svg" width="1024" height="1024" viewBox="0 0 1024 1024" role="img" aria-label="plumego toolkit square ${variant}">
  <rect width="1024" height="1024" rx="168" fill="${bg}"/>
  <circle cx="512" cy="512" r="68" fill="${accent}"/>
  <g stroke="${fg}" stroke-width="28" stroke-linecap="round">
    <line x1="512" y1="512" x2="256" y2="312"/>
    <line x1="512" y1="512" x2="768" y2="312"/>
    <line x1="512" y1="512" x2="828" y2="572"/>
    <line x1="512" y1="512" x2="636" y2="798"/>
    <line x1="512" y1="512" x2="312" y2="742"/>
  </g>
  <g fill="${muted}" stroke="${fg}" stroke-width="18">
    <circle cx="256" cy="312" r="44"/>
    <circle cx="768" cy="312" r="44"/>
    <circle cx="828" cy="572" r="44"/>
    <circle cx="636" cy="798" r="44"/>
    <circle cx="312" cy="742" r="44"/>
  </g>
</svg>
SVG
}

toolkit_horizontal() {
  local variant="$1"
  local bg fg accent muted
  bg="$(variant_bg "$variant" "#E5E7EB" "#111827")"
  fg="$(variant_fg "$variant" "#111827")"
  accent="$(variant_fg "$variant" "#0F766E")"
  muted="$(variant_fg "$variant" "#4B5563")"

  cat > "$ROOT/toolkit/svg/toolkit-horizontal-${variant}.svg" <<SVG
<svg xmlns="http://www.w3.org/2000/svg" width="1600" height="600" viewBox="0 0 1600 600" role="img" aria-label="plumego toolkit horizontal ${variant}">
  <rect width="1600" height="600" rx="72" fill="${bg}"/>
  <circle cx="280" cy="300" r="44" fill="${accent}"/>
  <g stroke="${fg}" stroke-width="16" stroke-linecap="round">
    <line x1="280" y1="300" x2="148" y2="188"/>
    <line x1="280" y1="300" x2="404" y2="186"/>
    <line x1="280" y1="300" x2="452" y2="314"/>
    <line x1="280" y1="300" x2="366" y2="446"/>
    <line x1="280" y1="300" x2="152" y2="430"/>
  </g>
  <g fill="${muted}" stroke="${fg}" stroke-width="10">
    <circle cx="148" cy="188" r="24"/>
    <circle cx="404" cy="186" r="24"/>
    <circle cx="452" cy="314" r="24"/>
    <circle cx="366" cy="446" r="24"/>
    <circle cx="152" cy="430" r="24"/>
  </g>
  <text x="560" y="360" fill="${fg}" font-size="176" font-family="Avenir Next, SF Pro Display, Segoe UI, Arial, sans-serif" font-weight="650" letter-spacing="-2">plumego</text>
</svg>
SVG
}

wordmark_square() {
  local variant="$1"
  local bg fg accent
  bg="$(variant_bg "$variant" "#FFFFFF" "#0B1F3A")"
  fg="$(variant_fg "$variant" "#0B1F3A")"
  accent="$(variant_fg "$variant" "#2DD4BF")"

  cat > "$ROOT/wordmark/svg/wordmark-square-${variant}.svg" <<SVG
<svg xmlns="http://www.w3.org/2000/svg" width="1024" height="1024" viewBox="0 0 1024 1024" role="img" aria-label="plumego wordmark square ${variant}">
  <rect width="1024" height="1024" rx="168" fill="${bg}"/>
  <text x="122" y="556" fill="${fg}" font-size="228" font-family="Avenir Next, SF Pro Display, Segoe UI, Arial, sans-serif" font-weight="700" letter-spacing="-6">plume</text>
  <text x="590" y="556" fill="${accent}" font-size="228" font-family="Avenir Next, SF Pro Display, Segoe UI, Arial, sans-serif" font-weight="800" letter-spacing="-8">go</text>
  <path d="M128 636H896" stroke="${fg}" stroke-opacity="0.2" stroke-width="20" stroke-linecap="round"/>
</svg>
SVG
}

wordmark_horizontal() {
  local variant="$1"
  local bg fg accent
  bg="$(variant_bg "$variant" "#FFFFFF" "#0B1F3A")"
  fg="$(variant_fg "$variant" "#0B1F3A")"
  accent="$(variant_fg "$variant" "#2DD4BF")"

  cat > "$ROOT/wordmark/svg/wordmark-horizontal-${variant}.svg" <<SVG
<svg xmlns="http://www.w3.org/2000/svg" width="1600" height="600" viewBox="0 0 1600 600" role="img" aria-label="plumego wordmark horizontal ${variant}">
  <rect width="1600" height="600" rx="72" fill="${bg}"/>
  <text x="132" y="370" fill="${fg}" font-size="260" font-family="Avenir Next, SF Pro Display, Segoe UI, Arial, sans-serif" font-weight="700" letter-spacing="-8">plume</text>
  <text x="736" y="370" fill="${accent}" font-size="260" font-family="Avenir Next, SF Pro Display, Segoe UI, Arial, sans-serif" font-weight="850" letter-spacing="-10">go</text>
  <path d="M136 436H1468" stroke="${fg}" stroke-opacity="0.22" stroke-width="14" stroke-linecap="round"/>
</svg>
SVG
}

generate_svg() {
  for variant in color bw reverse; do
    monogram_square "$variant"
    monogram_horizontal "$variant"
    toolkit_square "$variant"
    toolkit_horizontal "$variant"
    wordmark_square "$variant"
    wordmark_horizontal "$variant"
  done
}

assemble_final_svg() {
  cp "$ROOT/wordmark/svg/wordmark-horizontal-color.svg" "$ROOT/final/svg/plumego-primary-horizontal-light.svg"
  cp "$ROOT/wordmark/svg/wordmark-horizontal-reverse.svg" "$ROOT/final/svg/plumego-primary-horizontal-dark.svg"
  cp "$ROOT/wordmark/svg/wordmark-square-color.svg" "$ROOT/final/svg/plumego-primary-square-light.svg"
  cp "$ROOT/wordmark/svg/wordmark-square-reverse.svg" "$ROOT/final/svg/plumego-primary-square-dark.svg"

  cp "$ROOT/monogram/svg/monogram-horizontal-color.svg" "$ROOT/final/svg/plumego-backup-horizontal-light.svg"
  cp "$ROOT/monogram/svg/monogram-horizontal-reverse.svg" "$ROOT/final/svg/plumego-backup-horizontal-dark.svg"
  cp "$ROOT/monogram/svg/monogram-square-color.svg" "$ROOT/final/svg/plumego-backup-square-light.svg"
  cp "$ROOT/monogram/svg/monogram-square-reverse.svg" "$ROOT/final/svg/plumego-backup-square-dark.svg"
}

generate_svg
assemble_final_svg
python3 "$ROOT/render_png.py"

echo "Brand assets generated under: $ROOT"
