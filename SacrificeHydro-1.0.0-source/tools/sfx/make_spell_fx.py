#!/usr/bin/env python3
"""Textures for Hydro's spells (mods/hydro/fx):
  waterfall.png  Tidal Wall: 4x4 sheet of 16 frames of pouring water (laid on the Firewall's wall mesh)
  coral.png      Tidal Wall: a clump of reef coral on rock, a sprite set along the foot of the wall
  spring.png     Healing Spring: a soft blue zone on the ground
  dome.png       Geyser: the dome of water that swells up before it bursts
usage: make_spell_fx.py <fx dir>   (needs numpy and Pillow)"""
import os, sys
import numpy as np
from PIL import Image, ImageDraw, ImageFilter

fx = sys.argv[1]
rng = np.random.default_rng(5)

def save(arr, name):
    Image.fromarray((np.clip(arr, 0, 1) * 255).astype(np.uint8)).save(os.path.join(fx, name))

# --- waterfall.png ---
C = 256
sheet = np.zeros((4 * C, 4 * C, 4))
yy, xx = np.mgrid[0:C, 0:C] / C
streaks = []
for k in range(60):
    streaks.append((rng.uniform(0, 1), rng.uniform(0.004, 0.018), rng.uniform(0.4, 1.0), rng.uniform(0, 1), rng.uniform(2.5, 5.0)))
for f in range(16):
    t = f / 16
    col = np.zeros((C, C, 4))
    base = 0.35 + 0.15 * np.sin(2 * np.pi * (xx * 3 + yy * 0.5))
    col[..., 0] = 0.18; col[..., 1] = 0.5; col[..., 2] = 0.8; col[..., 3] = base
    for (x0, w, a, ph, freq) in streaks:  # white streaks running down, tileable left-right and in time
        dx = np.minimum(abs(xx - x0), 1 - abs(xx - x0))
        line = np.exp(-(dx / w) ** 2)
        pulse = 0.5 + 0.5 * np.sin(2 * np.pi * (freq * yy - t * round(freq) * 2 + ph))
        v = line * pulse ** 3 * a
        col[..., :3] += v[..., None] * np.array([0.8, 0.5, 0.2])
        col[..., 3] = np.maximum(col[..., 3], v)
    foot = np.clip((yy - 0.85) / 0.15, 0, 1) * (0.6 + 0.4 * np.sin(2 * np.pi * (xx * 9 + t)) ** 2)  # foam at the foot
    col[..., :3] += foot[..., None] * 0.6
    col[..., 3] = np.maximum(col[..., 3], foot)
    col[..., 3] *= np.clip(yy / 0.12, 0, 1)  # soft top edge
    r, c = divmod(f, 4)
    sheet[r * C:(r + 1) * C, c * C:(c + 1) * C] = col
save(sheet, 'waterfall.png')

# --- coral.png ---
S = 256
img = Image.new('RGBA', (S, S), (0, 0, 0, 0))
d = ImageDraw.Draw(img)
d.ellipse([20, 190, 236, 262], fill=(88, 84, 78, 255))  # rock
d.ellipse([50, 175, 170, 240], fill=(104, 98, 90, 255))
colors = [(214, 92, 74), (232, 140, 70), (210, 96, 150), (240, 190, 120)]
def branch(x, y, ang, length, width, col, depth):
    x2 = x + length * np.cos(ang); y2 = y - length * np.sin(ang)
    d.line([(x, y), (x2, y2)], fill=col + (255,), width=int(width))
    d.ellipse([x2 - width / 2, y2 - width / 2, x2 + width / 2, y2 + width / 2], fill=col + (255,))
    if depth > 0:
        for da in (-0.45, 0.4):
            branch(x2, y2, ang + da + rng.uniform(-0.15, 0.15), length * rng.uniform(0.6, 0.8), max(3, width * 0.72), col, depth - 1)
for k in range(5):
    branch(rng.uniform(60, 196), 205, np.pi / 2 + rng.uniform(-0.5, 0.5), rng.uniform(40, 60), rng.uniform(12, 18), colors[k % 4], 3)
img = img.filter(ImageFilter.GaussianBlur(0.8))
arr = np.asarray(img).astype(float) / 255
arr[..., :3] *= (0.8 + 0.3 * rng.random((S, S)))[..., None]
save(arr, 'coral.png')

# --- spring.png: a soft blue zone (Healing Spring) and dome.png: the Geyser's swelling water ---
def noise(N,cell):
    n=rng.random((N//cell,N//cell)); return np.array(Image.fromarray((n*255).astype(np.uint8)).resize((N,N),Image.BICUBIC))/255
N=512
yy,xx=np.mgrid[0:N,0:N]/(N-1)*2-1; r=np.sqrt(xx**2+yy**2)
n=0.6*noise(N,32)+0.4*noise(N,8)
col=np.stack([0.20+0.15*n,0.55+0.2*n,0.95+0.05*n],-1)
alpha=1.7*(np.clip((1-r)/0.2,0,1)*(0.35+0.15*n)+np.exp(-((r-0.95)/0.04)**2)*0.35*(r<1.02))
save(np.dstack([col,alpha]),'spring.png')
c=noise(N,16); c2=noise(N,6)
caustic=np.clip(1-abs(c-0.5)*6,0,1)**2+0.6*np.clip(1-abs(c2-0.5)*7,0,1)**2
col=np.stack([0.15+0.6*caustic,0.45+0.5*caustic,0.75+0.25*caustic],-1)
alpha=np.clip((1-r)/0.15,0,1)*(0.55+0.35*caustic)
save(np.dstack([col,alpha]),'dome.png')
print('ok')
