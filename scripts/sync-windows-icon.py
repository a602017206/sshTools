#!/usr/bin/env python3
"""从 build/appicon.png 生成 Windows 用的 build/windows/icon.ico。"""

from pathlib import Path

from PIL import Image

ROOT = Path(__file__).resolve().parents[1]
SRC = ROOT / "build" / "appicon.png"
DST = ROOT / "build" / "windows" / "icon.ico"
SIZES = [(256, 256), (128, 128), (64, 64), (48, 48), (32, 32), (16, 16)]


def main() -> None:
    img = Image.open(SRC).convert("RGBA")
    frames = [img.resize(size, Image.Resampling.LANCZOS) for size in SIZES]
    frames[0].save(DST, format="ICO", sizes=SIZES, append_images=frames[1:])
    print(f"wrote {DST} ({DST.stat().st_size} bytes)")


if __name__ == "__main__":
    main()
