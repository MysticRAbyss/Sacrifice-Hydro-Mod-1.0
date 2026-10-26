#!/usr/bin/env python3
"""Draws the Hydro spell icons (64x64 RGBA PNG) for the spellbar: mods/hydro/spells/<tag>.png.
They are round like the other gods' spell icons: the picture on a black disc in a bronze ring
(icon_style.py).
usage: make_spell_icons.py <out dir>   (needs numpy and Pillow)"""
import os, sys
import numpy as np
from PIL import Image, ImageFilter
import icon_style

N = 256  # drawn large, scaled down to 64
yy, xx = np.mgrid[0:N, 0:N] / (N - 1) * 2 - 1  # -1..1
rng = np.random.default_rng(7)

def background():
    r = np.sqrt(xx**2 + yy**2)
    base = np.stack([0.01 + 0.02 * (1 - r), 0.02 + 0.06 * (1 - r), 0.05 + 0.16 * (1 - r)], -1)
    noise = rng.standard_normal((N // 8, N // 8))
    noise = np.array(Image.fromarray(((noise + 3) / 6 * 255).clip(0, 255).astype(np.uint8)).resize((N, N), Image.BICUBIC)) / 255.0
    return np.clip(base * (0.8 + 0.4 * noise[..., None]), 0, 1)

def blend(img, mask, color, alpha=1.0):
    m = np.clip(mask, 0, 1)[..., None] * alpha
    return img * (1 - m) + np.array(color) * m

def soft(mask, r):
    im = Image.fromarray((np.clip(mask, 0, 1) * 255).astype(np.uint8)).filter(ImageFilter.GaussianBlur(r))
    return np.array(im) / 255.0

def save(img, path):
    # the picture is drawn for the whole square; shrink it a little so its middle fills the disc
    im = Image.fromarray((np.clip(img, 0, 1) * 255).astype(np.uint8))
    s = int(N * 0.86)
    canvas = Image.new('RGB', (N, N), (3, 6, 14))
    canvas.paste(im.resize((s, s), Image.LANCZOS), (int(icon_style.CX - s / 2), int(icon_style.CY - s / 2)))
    pic = np.asarray(canvas).astype(float) / 255
    icon_style.save(icon_style.frame(pic), path)

def water_jet():
    img = background()
    # a stream from the lower left to the upper right, widening, with a frothy head
    u = (xx - yy) / np.sqrt(2)  # along the jet
    v = (xx + yy) / np.sqrt(2)  # across
    width = 0.06 + 0.12 * (u + 1) / 2
    core = np.exp(-(v / width) ** 2) * (u > -0.95) * (u < 0.55)
    ripple = 0.5 + 0.5 * np.sin(18 * u + 6 * v)
    img = blend(img, soft(core, 3), (0.15, 0.55, 0.85), 0.9)
    img = blend(img, soft(core * (np.abs(v) < width * 0.35) * ripple, 2), (0.75, 0.95, 1.0), 0.8)
    head = np.exp(-((u - 0.55) ** 2 + v ** 2) / 0.03)
    img = blend(img, soft(head, 4), (0.85, 0.97, 1.0), 0.9)
    for _ in range(40):  # spray drops around the head
        a = rng.uniform(0, 2 * np.pi); d = rng.uniform(0.1, 0.45)
        cx, cy = 0.39 + d * np.cos(a) * 0.7, -0.39 + d * np.sin(a) * 0.7
        img = blend(img, np.exp(-((xx - cx) ** 2 + (yy - cy) ** 2) / rng.uniform(0.0004, 0.0015)), (0.8, 0.95, 1.0), 0.9)
    return img

def bubble_armor():
    img = background()
    r = np.sqrt(xx**2 + (yy + 0.02) ** 2)
    # a hooded figure inside
    fig = ((xx / 0.22) ** 2 + ((yy - 0.25) / 0.45) ** 2 < 1) | ((xx / 0.12) ** 2 + ((yy + 0.28) / 0.13) ** 2 < 1)
    img = blend(img, soft(fig.astype(float), 3), (0.05, 0.12, 0.2), 0.95)
    shell = np.exp(-((r - 0.72) / 0.07) ** 2)
    img = blend(img, soft(r < 0.75, 2) * 1.0, (0.2, 0.6, 0.9), 0.25)
    img = blend(img, shell, (0.55, 0.88, 1.0), 0.85)
    hl = np.exp(-(((xx + 0.32) / 0.16) ** 2 + ((yy + 0.38) / 0.09) ** 2))
    img = blend(img, soft(hl, 2), (1, 1, 1), 0.9)
    for _ in range(12):
        cx, cy, s = rng.uniform(-0.6, 0.6), rng.uniform(-0.6, 0.6), rng.uniform(0.03, 0.07)
        if cx * cx + cy * cy > 0.45: continue
        rr = np.sqrt((xx - cx) ** 2 + (yy - cy) ** 2)
        img = blend(img, np.exp(-((rr - s) / 0.012) ** 2), (0.7, 0.95, 1.0), 0.7)
    return img

def undertow():
    img = background()
    r = np.sqrt(xx**2 + yy**2) + 1e-6
    a = np.arctan2(yy, xx)
    spiral = 0.5 + 0.5 * np.cos(4 * (a + 5.0 * np.log(r)))
    disc = np.clip((0.9 - r) / 0.15, 0, 1)
    img = blend(img, disc, (0.05, 0.25, 0.45), 0.9)
    img = blend(img, soft(spiral ** 6 * disc, 2), (0.55, 0.85, 1.0), 0.85)
    img = blend(img, np.exp(-(r / 0.16) ** 2), (0.0, 0.03, 0.08), 0.95)  # the dark eye
    # a hand reaching out of it
    hand = ((xx - 0.05) / 0.1) ** 2 + ((yy + 0.05) / 0.22) ** 2 < 1
    img = blend(img, soft(hand.astype(float), 2), (0.12, 0.2, 0.25), 0.0)
    return img

def healing_spring():
    img = background()
    r = np.sqrt(xx**2 + (yy - 0.45) ** 2 * 4)
    pool = np.clip((0.75 - r) / 0.1, 0, 1)
    img = blend(img, pool, (0.25, 0.8, 0.85), 0.85)
    rings = (0.5 + 0.5 * np.cos(r * 30)) ** 6 * pool
    img = blend(img, rings, (0.85, 1.0, 1.0), 0.6)
    # the fountain: an arc of water rising from the pool
    for k in range(-3, 4):
        a = k * 0.18
        tt = np.linspace(0, 1, 120)
        px = 0.55 * np.sin(a) * tt * 2; py = 0.45 - 1.1 * tt + 0.9 * tt**2
        for x0, y0 in zip(px[::3], py[::3]):
            img = blend(img, np.exp(-((xx - x0) ** 2 + (yy - y0) ** 2) / 0.0012), (0.8, 1.0, 1.0), 0.7)
    glow = np.exp(-(xx**2 + (yy + 0.2) ** 2) / 0.08)
    img = blend(img, glow, (0.9, 1.0, 0.7), 0.5)  # healing light
    plus = ((abs(xx) < 0.05) & (abs(yy + 0.2) < 0.18)) | ((abs(yy + 0.2) < 0.05) & (abs(xx) < 0.18))
    img = blend(img, soft(plus.astype(float), 2), (1.0, 1.0, 0.85), 0.9)
    return img

def geyser():
    img = background()
    ground = soft((yy > 0.55).astype(float), 3)
    img = blend(img, ground, (0.22, 0.16, 0.1), 0.95)
    col = np.exp(-(xx / (0.12 + 0.18 * (1 - (yy + 1) / 2))) ** 2) * (yy > -0.85) * (yy < 0.65)
    img = blend(img, soft(col, 3), (0.3, 0.7, 0.95), 0.9)
    img = blend(img, soft(col * (abs(xx) < 0.05), 2), (0.9, 1.0, 1.0), 0.8)
    top = np.exp(-(xx**2 + (yy + 0.75) ** 2) / 0.05)
    img = blend(img, soft(top, 3), (0.85, 0.97, 1.0), 0.9)
    for _ in range(50):
        a = rng.uniform(-np.pi, 0); d = rng.uniform(0.2, 0.75)
        cx, cy = d * np.cos(a), -0.75 + d * np.sin(a) * 0.6 + d * 0.8
        img = blend(img, np.exp(-((xx - cx) ** 2 + (yy - cy) ** 2) / 0.0009), (0.8, 0.95, 1.0), 0.85)
    return img

def tidal_wall():
    img = background()
    wall = (yy > -0.7) & (yy < 0.55)
    streak = 0.5 + 0.5 * np.sin(xx * 40 + np.sin(yy * 6) * 2)
    img = blend(img, soft(wall.astype(float), 3), (0.15, 0.5, 0.8), 0.85)
    img = blend(img, soft(wall * streak ** 5, 1), (0.85, 0.97, 1.0), 0.6)
    crest = np.exp(-((yy + 0.7) / 0.06) ** 2) * (0.6 + 0.4 * np.sin(xx * 25) ** 2)
    img = blend(img, crest, (0.95, 1.0, 1.0), 0.9)
    for k in range(7):  # coral at the foot
        cx = -0.8 + k * 0.27
        coral = np.exp(-(((xx - cx) / 0.08) ** 2 + ((yy - 0.66) / 0.14) ** 2))
        img = blend(img, soft(coral, 2), [(0.85, 0.35, 0.3), (0.95, 0.55, 0.3), (0.85, 0.4, 0.6)][k % 3], 0.95)
    return img

def acid_rain():
    img = background()
    cloud = np.zeros_like(xx)
    for cx, cy, rr in ((-0.35, -0.55, 0.32), (0.05, -0.62, 0.38), (0.42, -0.5, 0.3), (0.0, -0.4, 0.3)):
        cloud = np.maximum(cloud, np.clip((rr - np.sqrt((xx - cx) ** 2 + (yy - cy) ** 2)) / 0.08, 0, 1))
    img = blend(img, soft(cloud, 2), (0.25, 0.3, 0.22), 0.95)
    for k in range(18):
        x0 = rng.uniform(-0.75, 0.75); y0 = rng.uniform(-0.2, 0.55)
        drop = np.exp(-(((xx - x0) / 0.025) ** 2 + ((yy - y0) / 0.11) ** 2))
        img = blend(img, drop, (0.6, 0.95, 0.25), 0.9)
    puddle = np.exp(-((yy - 0.78) / 0.06) ** 2) * (abs(xx) < 0.8)
    img = blend(img, soft(puddle, 2), (0.45, 0.75, 0.2), 0.7)
    return img

def _draw(img, paint):  # draws with Pillow on a 256x256 layer and blends it over the picture
    from PIL import ImageDraw
    layer = Image.new('RGBA', (N, N), (0, 0, 0, 0))
    paint(ImageDraw.Draw(layer))
    a = np.asarray(layer.filter(ImageFilter.GaussianBlur(1.2))).astype(float) / 255
    return img * (1 - a[..., 3:]) + a[..., :3] * a[..., 3:]

def tsunami():
    img = background()
    def paint(d):
        # the sea, then a wave rising on the right and curling over to the left
        d.rectangle([0, 190, 256, 256], fill=(20, 80, 140, 255))
        wave = [(256, 200), (256, 60), (225, 30), (180, 22), (135, 35), (100, 62), (82, 92), (90, 112), (112, 118), (126, 104),
                (122, 86), (140, 72), (165, 80), (178, 110), (170, 150), (150, 185), (120, 200)]
        d.polygon(wave, fill=(30, 110, 180, 255))
        d.polygon([(256, 200), (256, 90), (215, 70), (190, 110), (185, 160), (165, 200)], fill=(20, 70, 130, 255))
        d.line([(256, 60), (225, 30), (180, 22), (135, 35), (100, 62), (82, 92), (90, 112), (112, 118), (126, 104)], fill=(235, 250, 255, 255), width=7)
        for k in range(28):
            x, y = rng.uniform(40, 130), rng.uniform(40, 150)
            r = rng.uniform(2, 5)
            d.ellipse([x - r, y - r, x + r, y + r], fill=(220, 245, 255, 230))
        d.line([(0, 192), (256, 192)], fill=(200, 235, 255, 200), width=4)
    return _draw(img, paint)

def kraken():
    img = background()
    def paint(d):
        # the whirlpool
        for k in range(5):
            r = 40 + k * 18
            d.ellipse([128 - r * 1.4, 200 - r * 0.45, 128 + r * 1.4, 200 + r * 0.45], outline=(90, 170, 220, 200 - 30 * k), width=4)
        # four tentacles rising and curling
        for k, (x0, curl) in enumerate(((50, -1), (100, 1), (156, -1), (206, 1))):
            pts = []
            for i in range(30):
                t = i / 29
                x = x0 + curl * 30 * np.sin(t * 2.6) * t
                y = 205 - 175 * t + 25 * t ** 3
                pts.append((x, y))
            for i in range(29):
                w = int(22 * (1 - i / 29) + 4)
                d.line([pts[i], pts[i + 1]], fill=(150, 45, 90, 255), width=w)
            for i in range(3, 26, 4):  # suckers
                x, y = pts[i]; r = 4 * (1 - i / 29) + 2
                d.ellipse([x + curl * 4 - r, y - r, x + curl * 4 + r, y + r], fill=(235, 190, 190, 255))
    return _draw(img, paint)

if __name__ == '__main__':
    out = sys.argv[1]
    os.makedirs(out, exist_ok=True)
    save(water_jet(), os.path.join(out, 'wjet.png'))
    save(bubble_armor(), os.path.join(out, 'bbar.png'))
    save(undertow(), os.path.join(out, 'undt.png'))
    save(healing_spring(), os.path.join(out, 'hspr.png'))
    save(geyser(), os.path.join(out, 'gysr.png'))
    save(tidal_wall(), os.path.join(out, 'tdwl.png'))
    save(acid_rain(), os.path.join(out, 'acrn.png'))
    save(tsunami(), os.path.join(out, 'tsnm.png'))
    save(kraken(), os.path.join(out, 'krkn.png'))
