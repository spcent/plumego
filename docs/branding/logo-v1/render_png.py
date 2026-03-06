#!/usr/bin/env python3
from pathlib import Path
from PIL import Image, ImageDraw, ImageFont
import shutil

ROOT = Path(__file__).resolve().parent
SIZES = [1024, 512, 256, 128]


def font(size: int, bold: bool = False):
    cands = [
        "/System/Library/Fonts/Supplemental/Arial Bold.ttf" if bold else "/System/Library/Fonts/Supplemental/Arial.ttf",
        "/System/Library/Fonts/Supplemental/Helvetica.ttc",
    ]
    for c in cands:
        try:
            return ImageFont.truetype(c, size)
        except OSError:
            pass
    return ImageFont.load_default()


def pal_monogram(v):
    return {
        "color": ("#FFFFFF", "#1F2937", "#00ADD8"),
        "bw": ("#FFFFFF", "#111111", "#111111"),
        "reverse": ("#0B1220", "#FFFFFF", "#FFFFFF"),
    }[v]


def pal_toolkit(v):
    return {
        "color": ("#E5E7EB", "#111827", "#0F766E", "#4B5563"),
        "bw": ("#FFFFFF", "#111111", "#111111", "#444444"),
        "reverse": ("#111827", "#FFFFFF", "#FFFFFF", "#FFFFFF"),
    }[v]


def pal_wordmark(v):
    return {
        "color": ("#FFFFFF", "#0B1F3A", "#2DD4BF"),
        "bw": ("#FFFFFF", "#111111", "#111111"),
        "reverse": ("#0B1F3A", "#FFFFFF", "#FFFFFF"),
    }[v]


def wh(layout, size):
    return (size, size) if layout == "square" else (size, int(round(size * 600 / 1600)))


def draw_monogram(layout, variant, size, out):
    bg, fg, ac = pal_monogram(variant)
    w, h = wh(layout, size)
    img = Image.new("RGBA", (w, h), bg)
    d = ImageDraw.Draw(img)
    if layout == "square":
        s = size / 1024
        d.rounded_rectangle((168*s,168*s,856*s,856*s), radius=140*s, outline=ac, width=max(2, int(34*s)))
        d.line((362*s,268*s,362*s,756*s), fill=fg, width=max(2, int(68*s)))
        d.arc((362*s,318*s,750*s,630*s), start=270, end=90, fill=fg, width=max(2, int(68*s)))
        d.line((362*s,318*s,570*s,318*s), fill=fg, width=max(2, int(68*s)))
        d.line((362*s,630*s,570*s,630*s), fill=fg, width=max(2, int(68*s)))
        for cx, cy in [(624,473),(672,473),(648,432),(648,514),(707,432)]:
            r = 20*s
            d.ellipse((cx*s-r, cy*s-r, cx*s+r, cy*s+r), fill=bg)
    else:
        sx, sy = w/1600, h/600
        d.rounded_rectangle((92*sx,92*sy,508*sx,508*sy), radius=92*sy, outline=ac, width=max(2, int(24*sy)))
        d.line((196*sx,150*sy,196*sx,450*sy), fill=fg, width=max(2, int(38*sy)))
        d.arc((196*sx,182*sy,448*sx,374*sy), start=270, end=90, fill=fg, width=max(2, int(38*sy)))
        d.line((196*sx,182*sy,328*sx,182*sy), fill=fg, width=max(2, int(38*sy)))
        d.line((196*sx,374*sy,328*sx,374*sy), fill=fg, width=max(2, int(38*sy)))
        for cx, cy in [(368,274),(396,274),(382,250),(382,298),(416,250)]:
            r = 12*sy
            d.ellipse((cx*sx-r, cy*sy-r, cx*sx+r, cy*sy+r), fill=bg)
        d.text((592*sx, 348*sy), "plumego", fill=fg, font=font(max(16, int(176*sy)), True), anchor="ls")
    img.save(out)


def draw_toolkit(layout, variant, size, out):
    bg, fg, ac, md = pal_toolkit(variant)
    w, h = wh(layout, size)
    img = Image.new("RGBA", (w, h), bg)
    d = ImageDraw.Draw(img)
    if layout == "square":
        s = size/1024
        d.ellipse((444*s,444*s,580*s,580*s), fill=ac)
        for (x1,y1),(x2,y2) in [((512,512),(256,312)),((512,512),(768,312)),((512,512),(828,572)),((512,512),(636,798)),((512,512),(312,742))]:
            d.line((x1*s,y1*s,x2*s,y2*s), fill=fg, width=max(2, int(28*s)))
        for cx,cy in [(256,312),(768,312),(828,572),(636,798),(312,742)]:
            r=44*s
            d.ellipse((cx*s-r,cy*s-r,cx*s+r,cy*s+r), fill=md, outline=fg, width=max(2, int(18*s)))
    else:
        sx, sy = w/1600, h/600
        d.ellipse((236*sx,256*sy,324*sx,344*sy), fill=ac)
        for (x1,y1),(x2,y2) in [((280,300),(148,188)),((280,300),(404,186)),((280,300),(452,314)),((280,300),(366,446)),((280,300),(152,430))]:
            d.line((x1*sx,y1*sy,x2*sx,y2*sy), fill=fg, width=max(2, int(16*sy)))
        for cx,cy in [(148,188),(404,186),(452,314),(366,446),(152,430)]:
            r=24*sy
            d.ellipse((cx*sx-r,cy*sy-r,cx*sx+r,cy*sy+r), fill=md, outline=fg, width=max(2, int(10*sy)))
        d.text((560*sx,360*sy), "plumego", fill=fg, font=font(max(16, int(176*sy)), True), anchor="ls")
    img.save(out)


def draw_wordmark(layout, variant, size, out):
    bg, fg, ac = pal_wordmark(variant)
    w, h = wh(layout, size)
    img = Image.new("RGBA", (w, h), bg)
    d = ImageDraw.Draw(img)
    if layout == "square":
        s=size/1024
        d.text((122*s,556*s), "plume", fill=fg, font=font(max(18, int(228*s)), True), anchor="ls")
        d.text((590*s,556*s), "go", fill=ac, font=font(max(18, int(228*s)), True), anchor="ls")
        d.line((128*s,636*s,896*s,636*s), fill=fg+"38", width=max(2, int(20*s)))
    else:
        sx, sy = w/1600, h/600
        d.text((132*sx,370*sy), "plume", fill=fg, font=font(max(16, int(260*sy)), True), anchor="ls")
        d.text((736*sx,370*sy), "go", fill=ac, font=font(max(16, int(260*sy)), True), anchor="ls")
        d.line((136*sx,436*sy,1468*sx,436*sy), fill=fg+"38", width=max(2, int(14*sy)))
    img.save(out)


def render_all():
    for design in ["monogram", "toolkit", "wordmark"]:
        out_dir = ROOT / design / "png"
        out_dir.mkdir(parents=True, exist_ok=True)
        for layout in ["square", "horizontal"]:
            for variant in ["color", "bw", "reverse"]:
                stem = f"{design}-{layout}-{variant}"
                for size in SIZES:
                    out = out_dir / f"{stem}-{size}.png"
                    if design == "monogram":
                        draw_monogram(layout, variant, size, out)
                    elif design == "toolkit":
                        draw_toolkit(layout, variant, size, out)
                    else:
                        draw_wordmark(layout, variant, size, out)


def copy_final_png():
    mapping = {
        "plumego-primary-horizontal-light": ("wordmark", "wordmark-horizontal-color"),
        "plumego-primary-horizontal-dark": ("wordmark", "wordmark-horizontal-reverse"),
        "plumego-primary-square-light": ("wordmark", "wordmark-square-color"),
        "plumego-primary-square-dark": ("wordmark", "wordmark-square-reverse"),
        "plumego-backup-horizontal-light": ("monogram", "monogram-horizontal-color"),
        "plumego-backup-horizontal-dark": ("monogram", "monogram-horizontal-reverse"),
        "plumego-backup-square-light": ("monogram", "monogram-square-color"),
        "plumego-backup-square-dark": ("monogram", "monogram-square-reverse"),
    }
    out_dir = ROOT / "final" / "png"
    out_dir.mkdir(parents=True, exist_ok=True)
    for target, (design, source) in mapping.items():
        for size in SIZES:
            shutil.copy2(ROOT / design / "png" / f"{source}-{size}.png", out_dir / f"{target}-{size}.png")


if __name__ == "__main__":
    render_all()
    copy_final_png()
    print(f"PNG assets generated under: {ROOT}")
