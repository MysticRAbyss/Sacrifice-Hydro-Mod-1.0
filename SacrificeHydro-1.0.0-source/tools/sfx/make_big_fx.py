#!/usr/bin/env python3
"""Textures for Hydro's last three spells (mods/hydro/fx):
  acidpool.png   Acid Rain: the stained, steaming ground under the cloud
  acid.png       Acid Rain: one falling drop (a narrow streak)
  acidcloud.png  Acid Rain: a puff of the dark, sickly cloud
  wave.png       Tsunami: the wave's water, foot to crest (v) and along the front (u, tiles)
  tentacle.png   Kraken: tentacle skin, around (u) and along (v, tiles)
usage: make_big_fx.py <fx dir>   (needs numpy and Pillow)"""
import os, sys
import numpy as np
from PIL import Image
fx = sys.argv[1]
rng = np.random.default_rng(21)
def save(a, n): Image.fromarray((np.clip(a, 0, 1) * 255).astype(np.uint8)).save(os.path.join(fx, n))
def noise(N, cell, M=None):
    M = M or N
    n = rng.random((max(2, M // cell), max(2, N // cell)))
    return np.array(Image.fromarray((n * 255).astype(np.uint8)).resize((N, M), Image.BICUBIC)) / 255
N = 512
yy, xx = np.mgrid[0:N, 0:N] / (N - 1) * 2 - 1
r = np.sqrt(xx**2 + yy**2)
n = 0.6 * noise(N, 32) + 0.4 * noise(N, 8)
col = np.stack([0.35 + 0.2 * n, 0.55 + 0.25 * n, 0.12 + 0.08 * n], -1)
alpha = np.clip((1 - r) / 0.25, 0, 1) * (0.45 + 0.35 * n)
save(np.dstack([col, alpha]), 'acidpool.png')
# a drop: a thin bright-green streak
S = 64
g = np.mgrid[0:S * 4, 0:S]
yy2, xx2 = g[0] / (S * 4 - 1), g[1] / (S - 1)
streak = np.exp(-((xx2 - 0.5) / 0.18) ** 2) * np.clip(yy2 * 1.4, 0, 1) * np.clip((1 - yy2) * 6, 0, 1)
col = np.stack([0.6 + 0.3 * streak, 0.95 * np.ones_like(streak), 0.25 + 0.2 * streak], -1)
save(np.dstack([col, streak * 0.9]), 'acid.png')
# a cloud puff: dark grey-green, soft edges
S = 256
yy3, xx3 = np.mgrid[0:S, 0:S] / (S - 1) * 2 - 1
r3 = np.sqrt(xx3**2 + yy3**2)
n3 = 0.5 * noise(S, 32) + 0.35 * noise(S, 12) + 0.15 * noise(S, 4)
puff = np.clip((0.95 - r3 - 0.45 * (1 - n3)) / 0.35, 0, 1)
col = np.stack([0.22 + 0.15 * n3, 0.27 + 0.18 * n3, 0.18 + 0.1 * n3], -1)
save(np.dstack([col, puff * 0.92]), 'acidcloud.png')
# the wave: deep blue at the foot, lighter and streaked up the face, white foam over the crest
W, H = 256, 256
v, u = np.mgrid[0:H, 0:W] / np.array([H - 1.0, W - 1.0])[:, None, None]
nu = noise(W, 16, H)
streaks = 0.5 + 0.5 * np.sin(2 * np.pi * (u * 9 + 0.15 * np.sin(v * 12)))
depth = v  # 0 back foot, 1 front foot; crest around 0.5-0.6
crest = np.exp(-((v - 0.52) / 0.08) ** 2)
face = np.clip((v - 0.55) / 0.45, 0, 1)
col = np.stack([0.08 + 0.25 * face + 0.75 * crest, 0.28 + 0.35 * face + 0.65 * crest, 0.55 + 0.3 * face + 0.45 * crest], -1)
col += (0.12 * streaks * (0.3 + face))[..., None] + (0.08 * nu)[..., None]
foam_foot = np.exp(-((v - 0.97) / 0.04) ** 2) * (0.5 + 0.5 * nu)
col += foam_foot[..., None] * 0.7
alpha = np.clip(0.75 + 0.25 * crest + 0.2 * face, 0, 1) * np.clip(v / 0.08, 0, 1)
save(np.dstack([col, alpha]), 'wave.png')
# tentacle skin: u around the arm (0.5 = the sucker side), v along it from the root (0) to the tip (1);
# dark purple scales on the back, a pale sucker-lined belly, and a pale leaf-shaped club from v=0.72
T = 512
v, u = np.mgrid[0:T, 0:T] / (T - 1.0)
nt = noise(T, 16); nt2 = noise(T, 5)
club = np.clip((v - 0.68) / 0.08, 0, 1)
under = np.exp(-((u - 0.5) / (0.2 + 0.12 * club)) ** 2)
scales = 0.5 + 0.5 * np.sin(2 * np.pi * (u * 18 + np.floor(v * 60) * 0.5)) * np.sin(2 * np.pi * v * 60)
back = np.stack([0.36 + 0.1 * nt + 0.05 * scales, 0.14 + 0.05 * nt, 0.24 + 0.08 * nt], -1)
belly = np.stack([0.86 * np.ones_like(u), 0.66 + 0.06 * nt2, 0.62 + 0.05 * nt2], -1)
col = back * (1 - under[..., None]) + belly * under[..., None]
col = col * (1 - 0.35 * club[..., None]) + np.array([0.82, 0.62, 0.66]) * 0.35 * club[..., None]
k = 0
for row in np.linspace(0.05, 0.96, 20):  # rows of suckers, bigger on the club
    big = 1.0 + 1.6 * (row > 0.7)
    for side in (-0.07, 0.07) if row < 0.7 else (-0.13, 0.0, 0.13):
        d = np.sqrt(((u - 0.5 - side) / (0.045 * big)) ** 2 + ((v - row) / (0.016 * big)) ** 2)
        ring = np.exp(-((d - 0.75) / 0.22) ** 2); hole = np.exp(-(d / 0.4) ** 2)
        col = col * (1 - 0.6 * ring[..., None]) + np.array([0.95, 0.8, 0.78]) * 0.6 * ring[..., None]
        col = col * (1 - 0.5 * hole[..., None]) + np.array([0.45, 0.22, 0.25]) * 0.5 * hole[..., None]
    k += 1
col *= (0.85 + 0.25 * nt2)[..., None]
save(np.dstack([col, np.ones_like(u)]), 'tentacle.png')
print('ok')
