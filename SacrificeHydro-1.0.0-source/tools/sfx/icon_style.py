"""The round icon frame shared by Hydro's creature and spell icons, in the style of the other gods'
icons: a black disc with a dark blue glow, ringed by a bronze band lit from above with a thin blue
rim at its foot. Pictures are drawn at 256x256 and scaled down to 64x64 RGBA."""
import numpy as np
from PIL import Image

N = 256
CX, CY, R = 128.0, 130.0, 104.0  # the ring's centre and its inner radius (px at 256)
BAND = 12.0                      # the bronze band's width
yy, xx = np.mgrid[0:N, 0:N].astype(float)
_dx, _dy = xx - CX, yy - CY
r = np.sqrt(_dx ** 2 + _dy ** 2)
ang_up = -_dy / (r + 1e-6)  # 1 at the top of the ring, -1 at the bottom

def smooth(e0, e1, x):
    t = np.clip((x - e0) / (e1 - e0), 0, 1)
    return t * t * (3 - 2 * t)

def disc():
    """the inside of the ring: black with a faint deep blue glow, (rgb, alpha)"""
    g = np.clip(1 - r / R, 0, 1)
    rgb = np.stack([0.01 + 0.02 * g, 0.015 + 0.04 * g, 0.03 + 0.11 * g], -1)
    a = 1 - smooth(R - 1, R + 1, r)
    return rgb, a

def ring():
    """the bronze band, (rgb, alpha)"""
    t = (r - R) / BAND  # 0 inside edge .. 1 outside edge
    inside = smooth(-0.08, 0.08, t) * (1 - smooth(0.92, 1.08, t))
    profile = np.sin(np.clip(t, 0, 1) * np.pi)  # rounded: bright middle, dark edges
    light = 0.3 + 0.75 * (0.5 + 0.5 * ang_up)  # lit from above
    dark = np.array([0.20, 0.11, 0.03]); mid = np.array([0.70, 0.47, 0.14]); hi = np.array([1.0, 0.86, 0.45])
    v = np.clip(profile * light, 0, 1)[..., None]
    rgb = np.where(v < 0.5, dark + (mid - dark) * (v / 0.5), mid + (hi - mid) * ((v - 0.5) / 0.5))
    # a thin blue rim just inside the band, strongest at the bottom
    rim = np.exp(-((r - (R - 2.5)) / 1.8) ** 2) * (0.35 + 0.65 * np.clip(-ang_up + 0.3, 0, 1))
    return rgb, inside, rim

def frame(picture_rgb, picture_a=None, over=None):
    """picture_rgb: 256x256x3 drawn for the disc (clipped to it); picture_a its coverage (default 1);
    over: optional (rgb, alpha) drawn on top of the ring (things that stick out of it)"""
    drgb, da = disc()
    pa = np.ones((N, N)) if picture_a is None else picture_a
    rgb = drgb * (1 - pa[..., None]) + picture_rgb * pa[..., None]
    rim_col = np.array([0.25, 0.5, 1.0])
    rgb_rim, ra, rim = ring()
    rgb = rgb * (1 - rim[..., None] * 0.8) + rim_col * rim[..., None] * 0.8
    a = da
    rgb = rgb * (1 - ra[..., None]) + rgb_rim * ra[..., None]
    a = np.maximum(a, ra)
    if over is not None:
        orgb, oa = over
        rgb = rgb * (1 - oa[..., None]) + orgb * oa[..., None]
        a = np.maximum(a, oa)
    return np.dstack([np.clip(rgb, 0, 1), np.clip(a, 0, 1)])

def save(rgba, path, size=64):
    # premultiply before scaling so the transparent edge does not pick up dark fringes
    p = rgba.copy(); p[..., :3] *= p[..., 3:]
    im = Image.fromarray((np.clip(p, 0, 1) * 255).astype(np.uint8), 'RGBA').resize((size, size), Image.LANCZOS)
    q = np.asarray(im).astype(float) / 255
    q[..., :3] = np.where(q[..., 3:] > 0.004, q[..., :3] / np.maximum(q[..., 3:], 1e-3), 0)
    Image.fromarray((np.clip(q, 0, 1) * 255).astype(np.uint8), 'RGBA').save(path)
