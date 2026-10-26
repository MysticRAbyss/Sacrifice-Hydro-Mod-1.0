#!/usr/bin/env python3
"""Undertow's textures (mods/hydro/fx): undertow.png, the whirlpool with drifting seaweed (made
from whirlpool.png), and algae.png, a strand of kelp used for the seaweed around the target.
usage: make_undertow_fx.py <fx dir>   (needs numpy and Pillow)"""
import os, sys
import numpy as np
from PIL import Image, ImageDraw, ImageFilter

fx = sys.argv[1]
rng = np.random.default_rng(11)

# --- undertow.png: the whirlpool, darker and greener, with spiralling strands of algae ---
w = np.asarray(Image.open(os.path.join(fx, 'whirlpool.png')).convert('RGBA')).astype(float) / 255
N = w.shape[0]
rgb, a = w[..., :3], w[..., 3]
lum = rgb.mean(-1)
murky = np.stack([0.10 + 0.55 * lum, 0.32 + 0.60 * lum, 0.30 + 0.45 * lum], -1)  # sea-green water
yy, xx = np.mgrid[0:N, 0:N] / (N - 1) * 2 - 1
r = np.sqrt(xx**2 + yy**2) + 1e-6
ang = np.arctan2(yy, xx)
strands = np.zeros((N, N))
for k in range(7):  # strands of algae following the swirl
    phase = rng.uniform(0, 2 * np.pi)
    s = np.cos(ang + 4.2 * np.log(r) + phase)
    width = rng.uniform(0.04, 0.09)
    band = np.exp(-((1 - s) / width) ** 2) * (r > rng.uniform(0.15, 0.3)) * (r < rng.uniform(0.7, 0.95))
    strands = np.maximum(strands, band * rng.uniform(0.6, 1.0))
blotch = np.array(Image.fromarray((rng.random((N // 16, N // 16)) * 255).astype(np.uint8)).resize((N, N), Image.BICUBIC)) / 255
strands *= 0.6 + 0.6 * blotch
algae_col = np.array([0.16, 0.36, 0.10])
out = murky * (1 - strands[..., None] * 0.85) + algae_col * strands[..., None] * 0.85
alpha = np.clip(a + strands * 0.8 * (r < 0.97), 0, 1)
Image.fromarray((np.dstack([out, alpha]) * 255).clip(0, 255).astype(np.uint8)).save(os.path.join(fx, 'undertow.png'))

# --- algae.png: one wavy kelp strand with leaves, on a transparent square ---
S = 256
img = Image.new('RGBA', (S, S), (0, 0, 0, 0))
d = ImageDraw.Draw(img)
pts = []
for i in range(41):
    t = i / 40
    y = S * (0.97 - 0.94 * t)
    x = S * (0.5 + 0.12 * np.sin(t * 7.0) * t)
    pts.append((x, y))
for i in range(len(pts) - 1):  # the stem, thinner at the top
    wdt = int(14 * (1 - 0.7 * i / len(pts))) + 2
    d.line([pts[i], pts[i + 1]], fill=(58, 96, 34, 255), width=wdt)
for i in range(4, len(pts) - 2, 3):  # leaves alternating left and right
    x, y = pts[i]
    side = 1 if (i // 3) % 2 else -1
    L = rng.uniform(40, 62)
    leaf = [(x, y), (x + side * L * 0.5, y - L * 0.45), (x + side * L, y - L * 0.55), (x + side * L * 0.45, y - L * 0.25)]  # long, narrow kelp blades
    d.polygon(leaf, fill=(int(rng.uniform(70, 100)), int(rng.uniform(115, 145)), 40, 225))
img = img.filter(ImageFilter.GaussianBlur(1.2))
arr = np.asarray(img).astype(float)
shade = 0.75 + 0.35 * rng.random((S, S))  # a little mottling
arr[..., :3] *= shade[..., None]
Image.fromarray(arr.clip(0, 255).astype(np.uint8)).save(os.path.join(fx, 'algae.png'))
print('ok')
