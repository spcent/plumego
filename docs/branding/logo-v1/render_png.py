#!/usr/bin/env python3
from __future__ import annotations

import math
import shutil
from pathlib import Path
from PIL import Image, ImageDraw, ImageFont

ROOT = Path(__file__).resolve().parent
SIZES = [1024, 512, 256, 128]
DESIGNS = ["monogram", "toolkit", "wordmark"]
VARIANTS = ["color", "bw", "reverse"]
LAYOUTS = ["square", "horizontal"]


def get_font(px: int, bold: bool = False) -> ImageFont.ImageFont:
    candidates = []
    if bold:
        candidates.extend(
            [
                "/System/Library/Fonts/Supplemental/Arial Bold.ttf",
                "/System/Library/Fonts/Supplemental/Helvetica.ttc",
            ]
        )
    else:
        candidates.extend(
            [
                "/System/Library/Fonts/Supplemental/Arial.ttf",
                "/System/Library/Fonts/Supplemental/Helvetica.ttc",
            ]
        )
    for path in candidates:
        try:
            return ImageFont.truetype(path, px)
        except OSError:
            continue
    return ImageFont.load_default()


def palette_monogram(variant: str):
    if variant == "color":
        return "#FFFFFF", "#1F2937", "#00ADD8"
    if variant == "bw":
        return "#FFFFFF", "#111111", "#111111"
    return "#0B1220", "#FFFFFF", "#FFFFFF"


def palette_toolkit(variant: str):
    if variant == "color":
        return "#E5E7EB", "#111827", "#0F766E", "#4B5563"
    if variant == "bw":
        return "#FFFFFF", "#111111", "#111111", "#444444"
    return "#111827", "#FFFFFF", "#FFFFFF", "#FFFFFF"


def palette_wordmark(variant: str):
    if variant == "color":
        return "#FFFFFF", "#0B1F3A", "#2DD4BF"
    if variant == "bw":
        return "#FFFFFF", "#111111", "#111111"
    return "#0B1F3A", "#FFFFFF", "#FFFFFF"


def canvas(layout: str, size: int):
    if layout == "square":
        return size, size
    return size, max(1, int(round(size * 600 / 1600)))


def draw_monogram(layout: str, variant: str, size: int, out_path: Path):
    bg, fg, accent = palette_monogram(variant)
    w, h = canvas(layout, size)
    img = Image.new("RGBA", (w, h), bg)
    d = ImageDraw.Draw(img)

    if layout == "square":
        s = size / 1024
        d.rounded_rectangle((168 * s, 168 * s, 856 * s, 856 * s), radius=140 * s, outline=accent, width=max(2, int(34 * s)))
        d.line((362 * s, 268 * s, 362 * s, 756 * s), fill=fg, width=max(2, int(68 * s)))
        d.arc((362 * s, 318 * s, 750 * s, 630 * s), start=270, end=90, fill=fg, width=max(2, int(68 * s)))
        d.line((362 * s, 318 * s, 570 * s, 318 * s), fill=fg, width=max(2, int(68 * s)))
        d.line((362 * s, 630 * s, 570 * s, 630 * s), fill=fg, width=max(2, int(68 * s)))
        for cx, cy in [(624, 473), (672, 473), (648, 432), (648, 514), (707, 432)]:
            r = 20 * s
            d.ellipse((cx * s - r, cy * s - r, cx * s + r, cy * s + r), fill=bg)
    else:
        sx = w / 1600
        sy = h / 600
        d.rounded_rectangle((92 * sx, 92 * sy, 508 * sx, 508 * sy), radius=92 * sy, outline=accent, width=max(2, int(24 * sy)))
        d.line((196 * sx, 150 * sy, 196 * sx, 450 * sy), fill=fg, width=max(2, int(38 * sy)))
        d.arc((196 * sx, 182 * sy, 448 * sx, 374 * sy), start=270, end=90, fill=fg, width=max(2, int(38 * sy)))
        d.line((196 * sx, 182 * sy, 328 * sx, 182 * sy), fill=fg, width=max(2, int(38 * sy)))
        d.line((196 * sx, 374 * sy, 328 * sx, 374 * sy), fill=fg, width=max(2, int(38 * sy)))
        for cx, cy in [(368, 274), (396, 274), (382, 250), (382, 298), (416, 250)]:
            r = 12 * sy
            d.ellipse((cx * sx - r, cy * sy - r, cx * sx + r, cy * sy + r), fill=bg)
        font = get_font(max(16, int(176 * sy)), bold=True)
        d.text((592 * sx, 348 * sy), "plumego", fill=fg, font=font, anchor="ls")

    img.save(out_path)


def draw_toolkit(layout: str, variant: str, size: int, out_path: Path):
    bg, fg, accent, muted = palette_toolkit(variant)
    w, h = canvas(layout, size)
    img = Image.new("RGBA", (w, h), bg)
    d = ImageDraw.Draw(img)

    if layout == "square":
        s = size / 1024
        d.ellipse((444 * s, 444 * s, 580 * s, 580 * s), fill=accent)
        edges = [((512, 512), (256, 312)), ((512, 512), (768, 312)), ((512, 512), (828, 572)), ((512, 512), (636, 798)), ((512, 512), (312, 742))]
        for (x1, y1), (x2, y2) in edges:
            d.line((x1 * s, y1 * s, x2 * s, y2 * s), fill=fg, width=max(2, int(28 * s)))
        for cx, cy in [(256, 312), (768, 312), (828, 572), (636, 798), (312, 742)]:
            r = 44 * s
            d.ellipse((cx * s - r, cy * s - r, cx * s + r, cy * s + r), fill=muted, outline=fg, width=max(2, int(18 * s)))
    else:
        sx = w / 1600
        sy = h / 600
        d.ellipse((236 * sx, 256 * sy, 324 * sx, 344 * sy), fill=accent)
        edges = [((280, 300), (148, 188)), ((280, 300), (404, 186)), ((280, 300), (452, 314)), ((280, 300), (366, 446)), ((280, 300), (152, 430))]
        for (x1, y1), (x2, y2) in edges:
            d.line((x1 * sx, y1 * sy, x2 * sx, y2 * sy), fill=fg, width=max(2, int(16 * sy)))
        for cx, cy in [(148, 188), (404, 186), (452, 314), (366, 446), (152, 430)]:
            r = 24 * sy
            d.ellipse((cx * sx - r, cy * sy - r, cx * sx + r, cy * sy + r), fill=muted, outline=fg, width=max(2, int(10 * sy)))
        font = get_font(max(16, int(176 * sy)), bold=True)
        d.text((560 * sx, 360 * sy), "plumego", fill=fg, font=font, anchor="ls")

    img.save(out_path)


def draw_wordmark(layout: str, variant: str, size: int, out_path: Path):
    bg, fg, accent = palette_wordmark(variant)
    w, h = canvas(layout, size)
    img = Image.new("RGBA", (w, h), bg)
    d = ImageDraw.Draw(img)

    if layout == "square":
        s = size / 1024
        font1 = get_font(max(18, int(228 * s)), bold=True)
        font2 = get_font(max(18, int(228 * s)), bold=True)
        d.text((122 * s, 556 * s), "plume", fill=fg, font=font1, anchor="ls")
        d.text((590 * s, 556 * s), "go", fill=accent, font=font2, anchor="ls")
        d.line((128 * s, 636 * s, 896 * s, 636 * s), fill=fg + "38", width=max(2, int(20 * s)))
    else:
        sx = w / 1600
        sy = h / 600
        font1 = get_font(max(16, int(260 * sy)), bold=True)
        font2 = get_font(max(16, int(260 * sy)), bold=True)
        d.text((132 * sx, 370 * sy), "plume", fill=fg, font=font1, anchor="ls")
        d.text((736 * sx, 370 * sy), "go", fill=accent, font=font2, anchor="ls")
        d.line((136 * sx, 436 * sy, 1468 * sx, 436 * sy), fill=fg + "38", width=max(2, int(14 * sy)))

    img.save(out_path)


def render_all():
    for design in DESIGNS:
        png_dir = ROOT / design / "png"
        png_dir.mkdir(parents=True, exist_ok=True)
        for layout in LAYOUTS:
            for variant in VARIANTS:
                stem = f"{design}-{layout}-{variant}"
                for size in SIZES:
                    path = png_dir / f"{stem}-{size}.png"
                    if design == "monogram":
                        draw_monogram(layout, variant, size, path)
                    elif design == "toolkit":
                        draw_toolkit(layout, variant, size, path)
                    else:
                        draw_wordmark(layout, variant, size, path)


def assemble_final():
    final_png_dir = ROOT / "final" / "png"
    final_png_dir.mkdir(parents=True, exist_ok=True)

    mapping = {
        "plumego-primary-horizontal-light": "wordmark-horizontal-color",
        "plumego-primary-horizontal-dark": "wordmark-horizontal-reverse",
        "plumego-primary-square-light": "wordmark-square-color",
        "plumego-primary-square-dark": "wordmark-square-reverse",
        "plumego-backup-horizontal-light": "monogram-horizontal-color",
        "plumego-backup-horizontal-dark": "monogram-horizontal-reverse",
        "plumego-backup-square-light": "monogram-square-color",
        "plumego-backup-square-dark": "monogram-square-reverse",
    }

    for target_stem, source_stem in mapping.items():
        source_design = "wordmark" if source_stem.startswith("wordmark") else "monogram"
        for size in SIZES:
            src = ROOT / source_design / "png" / f"{source_stem}-{size}.png"
            dst = final_png_dir / f"{target_stem}-{size}.png"
            shutil.copy2(src, dst)


if __name__ == "__main__":
    render_all()
    assemble_final()
    print(f"PNG assets generated under: {ROOT}")
