#!/usr/bin/env python3
"""Bubble Armor's textures (mods/hydro/fx): bubble.png, one small bubble (sprite), and
bubbleskin.png, the film of the big bubble (4x2 sheet of 8 frames, laid on the Air Shield's sphere).
usage: make_bubble_fx.py <fx dir>   (needs numpy and Pillow)"""
import os, sys
import numpy as np
from PIL import Image
fx = sys.argv[1]
rng = np.random.default_rng(9)
def save(a, n): Image.fromarray((np.clip(a, 0, 1) * 255).astype(np.uint8)).save(os.path.join(fx, n))
# a small bubble: clear middle, bright rim, a highlight
S = 128
yy, xx = np.mgrid[0:S, 0:S] / (S - 1) * 2 - 1
r = np.sqrt(xx**2 + yy**2)
rim = np.exp(-((r - 0.82) / 0.09) ** 2)
hl = np.exp(-(((xx + 0.35) / 0.18) ** 2 + ((yy + 0.38) / 0.12) ** 2))
col = np.stack([0.75 + 0.25 * hl, 0.92 + 0.08 * hl, np.ones_like(r)], -1)
alpha = np.clip(0.12 * (r < 0.85) + 0.8 * rim + 0.9 * hl, 0, 1) * (r < 0.98)
save(np.dstack([col, alpha]), 'bubble.png')
# the film: soft iridescent bands drifting from frame to frame
C = 128
sheet = np.zeros((2 * C, 4 * C, 4))
yy, xx = np.mgrid[0:C, 0:C] / C
for f in range(8):
    t = f / 8
    w = np.sin(2 * np.pi * (xx * 2 + yy * 1.3 + t)) + 0.6 * np.sin(2 * np.pi * (xx * 3.7 - yy * 2.2 - 2 * t))
    hue = 0.5 + 0.5 * np.sin(w * 2.0)
    col = np.stack([0.35 + 0.2 * hue, 0.7 + 0.12 * np.sin(w * 2 + 2.1), 0.95 + 0.05 * np.sin(w * 2 + 4.2)], -1)  # sea-blue, a hint of shimmer
    alpha = 0.05 + 0.13 * (0.5 + 0.5 * np.sin(w * 3)) ** 3
    r_, c_ = divmod(f, 4)
    sheet[r_ * C:(r_ + 1) * C, c_ * C:(c_ + 1) * C] = np.dstack([col, alpha])
save(sheet, 'bubbleskin.png')
print('ok')
