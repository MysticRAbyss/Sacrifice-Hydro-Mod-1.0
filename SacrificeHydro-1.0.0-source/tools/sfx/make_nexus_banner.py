# Nexus Mods images for Sacrifice: Hydro, made only from the mod's own icons
import sys, math
from PIL import Image, ImageDraw, ImageFont, ImageFilter
root = sys.argv[1]; out = sys.argv[2]; subtitle = sys.argv[3] if len(sys.argv) > 3 else ""
W, H = 1600, 900
bg = Image.new("RGB", (W, H))
px = bg.load()
for y in range(H):
    for x in range(W):
        t = y / H; r = math.hypot(x - W / 2, y - H * 0.42) / (W * 0.7)
        k = max(0.0, 1 - r)
        px[x, y] = (int(4 + 10 * k), int(18 + 40 * k + 10 * (1 - t)), int(34 + 80 * k + 20 * (1 - t)))
img = bg.convert("RGBA")
# soft caustic waves
waves = Image.new("RGBA", (W, H), (0, 0, 0, 0)); d = ImageDraw.Draw(waves)
for i in range(40):
    y0 = 40 + i * 22
    pts = [(x, y0 + 8 * math.sin(x / 90 + i * 0.7)) for x in range(0, W + 20, 20)]
    d.line(pts, fill=(120, 200, 255, 10), width=3)
img = Image.alpha_composite(img, waves.filter(ImageFilter.GaussianBlur(2)))
d = ImageDraw.Draw(img)
serif = "/usr/share/fonts/truetype/dejavu/DejaVuSerif-Bold.ttf"
def text(y, s, size, fill):
    f = ImageFont.truetype(serif, size)
    w = d.textlength(s, font=f)
    glow = Image.new("RGBA", (W, H), (0, 0, 0, 0)); gd = ImageDraw.Draw(glow)
    gd.text(((W - w) / 2, y), s, font=f, fill=(60, 170, 255, 160))
    img.alpha_composite(glow.filter(ImageFilter.GaussianBlur(8)))
    d.text(((W - w) / 2, y), s, font=f, fill=fill)
text(70, "SACRIFICE: HYDRO", 96, (236, 214, 150, 255))
text(190, "A sixth god rises from the deep", 38, (190, 225, 245, 255))
if subtitle:
    text(245, subtitle, 30, (236, 214, 150, 255))
creatures = ["tidekin", "reefstalker", "mistwing", "mantisshrimp", "brineclaw", "tidecaller", "puffy", "reefbreaker", "abyssguardian", "leviathan"]
spells = ["wjet", "bbar", "undt", "hspr", "gysr", "tdwl", "acrn", "tsnm", "krkn"]
def row(names, path, y, size, labels=None):
    n = len(names); gap = 22; total = n * size + (n - 1) * gap; x = (W - total) // 2
    for i, nm in enumerate(names):
        ic = Image.open(path(nm)).convert("RGBA").resize((size, size), Image.LANCZOS)
        img.alpha_composite(ic, (x + i * (size + gap), y))
    return y + size
f = ImageFont.truetype(serif, 26)
def label(y, s):
    w = d.textlength(s, font=f); d.text(((W - w) / 2, y), s, font=f, fill=(150, 200, 230, 255))
label(320, "10 creatures")
row(creatures, lambda n: f"{root}/creatures/{n}/icon.png", 365, 118)
label(520, "9 spells")
row(spells, lambda n: f"{root}/spells/{n}.png", 565, 118)
label(740, "Needs the original Sacrifice (GOG / Steam)  -  runs on sacengine")
img.convert("RGB").save(out, quality=92)
