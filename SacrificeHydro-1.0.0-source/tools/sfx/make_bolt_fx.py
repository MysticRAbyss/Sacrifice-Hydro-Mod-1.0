#!/usr/bin/env python3
"""waterbolt.png (mods/hydro/fx): the water bolt the Tidecaller and Puffy shoot, a glowing ball of
swirling sea water with a bright core and a soft rim, drawn as a sprite (additive).
usage: make_bolt_fx.py <fx dir>   (needs numpy and Pillow)"""
import os, sys
import numpy as np
from PIL import Image

out = sys.argv[1]
N = 128
yy, xx = np.mgrid[0:N, 0:N] / (N - 1) * 2 - 1
r = np.sqrt(xx ** 2 + yy ** 2)
a = np.arctan2(yy, xx)
swirl = 0.5 + 0.5 * np.sin(3 * a + 9 * r)  # water winding around the core
ball = np.clip((0.78 - r) / 0.12, 0, 1)
core = np.exp(-(r / 0.28) ** 2)
glow = np.exp(-(r / 0.75) ** 2) * 0.45
rim = np.exp(-((r - 0.66) / 0.07) ** 2) * 0.6
hl = np.exp(-(((xx + 0.25) / 0.14) ** 2 + ((yy + 0.3) / 0.09) ** 2))  # a highlight
col = np.zeros((N, N, 3))
col += ball[..., None] * (np.array([0.10, 0.45, 0.85]) * (0.7 + 0.5 * swirl[..., None]))
col += core[..., None] * np.array([0.75, 0.95, 1.0])
col += rim[..., None] * np.array([0.5, 0.85, 1.0])
col += hl[..., None] * 0.9
alpha = np.clip(np.maximum(ball * (0.6 + 0.3 * swirl), glow) + core * 0.4, 0, 1)
img = np.dstack([np.clip(col + glow[..., None] * np.array([0.1, 0.4, 0.7]), 0, 1), alpha])
Image.fromarray((img * 255).astype(np.uint8), 'RGBA').save(os.path.join(out, 'waterbolt.png'))
print('ok')
