#!/usr/bin/env python3
from pathlib import Path
from PIL import Image

ROOT = Path(__file__).resolve().parent
src = ROOT / "final" / "png" / "plumego-primary-square-light-1024.png"
out = ROOT / "final" / "favicon"
out.mkdir(parents=True, exist_ok=True)

img = Image.open(src).convert("RGBA")

for size in [16, 32, 48, 64, 180, 192, 512]:
    resized = img.resize((size, size), Image.Resampling.LANCZOS)
    resized.save(out / f"favicon-{size}.png")

ico_sizes = [(16, 16), (32, 32), (48, 48), (64, 64)]
img.save(out / "favicon.ico", sizes=ico_sizes)
print(f"favicon assets written to: {out}")
