#!/usr/bin/env python3
from pathlib import Path
from PIL import Image

root = Path(__file__).resolve().parent
src = root / "final" / "png" / "plumego-primary-square-light-1024.png"
out = root / "final" / "favicon"
out.mkdir(parents=True, exist_ok=True)
img = Image.open(src).convert("RGBA")
for size in [16, 32, 48, 64, 180, 192, 512]:
    img.resize((size, size), Image.Resampling.LANCZOS).save(out / f"favicon-{size}.png")
img.save(out / "favicon.ico", sizes=[(16,16), (32,32), (48,48), (64,64)])
print(f"favicon assets written to: {out}")
