#!/usr/bin/env python3
"""Round spellbook icons for the Hydro creatures, in the style of the other gods' creature icons:
the creature as a gold-and-blue figure (lit gold from above, deep blue in its shadows) standing in
the bronze ring, overlapping it.
Each creature is rendered on black and on white by the creature tool (tools/kappagen, run here with
the view chosen for it below), so that its outline can be recovered.
usage: make_creature_icons.py <kappagen binary> <mods/hydro/creatures> <work dir>
       (writes <creature>/icon.png; needs numpy and Pillow)"""
import os, sys, glob, subprocess
import numpy as np
from PIL import Image, ImageFilter
import icon_style as st

kappagen, dest, raw = sys.argv[1], sys.argv[2], sys.argv[3]
# the view of each creature: yaw (radians, 0 = from the front) and pitch (looking down)
VIEWS = {'abyssguardian': (0.8, 0.35), 'brineclaw': (0.0, 0.6), 'leviathan': (0.8, 1.0), 'mantisshrimp': (0.8, 0.6),
         'mistwing': (0.3, 1.15), 'puffy': (0.0, 0.35), 'reefbreaker': (0.8, 0.35), 'reefstalker': (5.5, 0.35),
         'tidecaller': (0.0, 0.35), 'tidekin': (0.0, 0.35)}
os.makedirs(raw, exist_ok=True)
for n, (yaw, pitch) in VIEWS.items():
    subprocess.run([kappagen, '-only', n, '-iconraw', f'{dest}:{raw}:{yaw}:{pitch}'], check=True, stdout=subprocess.DEVNULL)

def figure(name):
    k = np.asarray(Image.open(os.path.join(raw, name + '_k.png')).convert('RGB')).astype(float) / 255
    w = np.asarray(Image.open(os.path.join(raw, name + '_w.png')).convert('RGB')).astype(float) / 255
    a = np.clip(1 - (w - k).mean(-1), 0, 1)
    col = np.where(a[..., None] > 0.02, k / np.maximum(a[..., None], 1e-3), 0)
    ys, xs = np.nonzero(a > 0.05)
    y0, y1, x0, x1 = ys.min(), ys.max() + 1, xs.min(), xs.max() + 1
    return col[y0:y1, x0:x1], a[y0:y1, x0:x1]

def place(col, a):
    """scales the figure to fill the ring, overlapping it a little, standing on its lower edge"""
    h, w = a.shape
    s = min(244.0 / h, 252.0 / w)
    W, H = max(1, int(w * s)), max(1, int(h * s))
    ci = Image.fromarray((np.clip(col, 0, 1) * 255).astype(np.uint8)).resize((W, H), Image.LANCZOS)
    ai = Image.fromarray((a * 255).astype(np.uint8)).resize((W, H), Image.LANCZOS)
    C = np.zeros((st.N, st.N, 3)); A = np.zeros((st.N, st.N))
    ox = int(st.CX - W / 2); oy = int(min(st.CY + st.R + st.BAND * 0.6 - H, st.N - H - 2))
    if H < 0.75 * W: oy = int(st.CY - H / 2)  # low, wide creatures sit in the middle of the ring
    oy = max(oy, 2)
    C[oy:oy + H, ox:ox + W] = np.asarray(ci).astype(float) / 255
    A[oy:oy + H, ox:ox + W] = np.asarray(ai).astype(float) / 255
    return C, A

def stylise(C, A):
    L = C @ np.array([0.3, 0.55, 0.15])
    m = A > 0.3
    lo, hi = np.percentile(L[m], 3), np.percentile(L[m], 97)
    L = np.clip((L - lo) / (hi - lo + 1e-6), 0, 1)
    # light from above: the upper side of every part catches gold
    Ai = Image.fromarray((A * 255).astype(np.uint8))
    up = np.asarray(Ai.transform(Ai.size, Image.AFFINE, (1, 0, 0, 0, 1, -7))).astype(float) / 255  # the shape moved down 7 px
    top_edge = np.clip(A - up, 0, 1)  # parts with nothing above them
    top_edge = np.asarray(Image.fromarray((top_edge * 255).astype(np.uint8)).filter(ImageFilter.GaussianBlur(2))).astype(float) / 255
    g = np.clip(st.smooth(0.3, 0.6, L) + 1.5 * top_edge, 0, 1)
    blue_lo, blue_hi = np.array([0.02, 0.04, 0.16]), np.array([0.16, 0.36, 0.95])
    gold_lo, gold_hi = np.array([0.55, 0.33, 0.04]), np.array([1.0, 0.86, 0.36])
    lb = np.clip(L / 0.6, 0, 1)[..., None]
    blue = blue_lo + (blue_hi - blue_lo) * lb
    lg = np.clip((L - 0.3) / 0.7, 0, 1)[..., None]
    gold = gold_lo + (gold_hi - gold_lo) * lg
    rgb = blue * (1 - g[..., None]) + gold * g[..., None]
    # a dark outline so the figure reads against the ring
    outline = np.asarray(Ai.filter(ImageFilter.MaxFilter(7))).astype(float) / 255
    out_a = np.maximum(A, outline * 0.9)
    rgb = rgb * (A / np.maximum(out_a, 1e-3))[..., None] + np.array([0.02, 0.02, 0.05]) * (1 - A / np.maximum(out_a, 1e-3))[..., None]
    return rgb, out_a

names = sorted(VIEWS)
for n in names:
    col, a = figure(n)
    C, A = place(col, a)
    rgb, oa = stylise(C, A)
    icon = st.frame(np.zeros((st.N, st.N, 3)), np.zeros((st.N, st.N)), over=(rgb, oa))
    st.save(icon, os.path.join(dest, n, 'icon.png'))
    print('icon', n)
