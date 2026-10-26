#!/usr/bin/env python3
"""Builds Hydro's manalith: three pillars of reef rock crowned with branching coral, on a rocky base.
Writes <out>/manalith.bin (the mesh, assembled into the game's model format by the engine,
hydromod.d addHydroManalith) and two textures: reefrock.png and coral.png.
Coordinates: metres, z up, about 19 m tall like the other gods' manaliths.
usage: make_manalith.py <out dir>   (needs numpy and Pillow)"""
import os, sys, struct
import numpy as np
from PIL import Image

out = sys.argv[1]
os.makedirs(out, exist_ok=True)
rng = np.random.default_rng(42)

verts = []   # (pos3, uv2)
faces = []   # (a, b, c, texture index)
def add_vert(p, uv):
    verts.append((np.asarray(p, float), np.asarray(uv, float)))
    return len(verts) - 1

def tube(path, radii, sides, tex, uscale=1.0, vscale=0.25, bumpy=0.0, cap_top=True):
    """a closed tube along path (list of points) with the given radius at each point"""
    path = [np.asarray(p, float) for p in path]
    rings = []
    vacc = 0.0
    prev_n = None
    for i, p in enumerate(path):
        t = path[min(i + 1, len(path) - 1)] - path[max(i - 1, 0)]
        t = t / (np.linalg.norm(t) + 1e-9)
        ref = np.array([0.0, 0.0, 1.0]) if abs(t[2]) < 0.9 else np.array([1.0, 0.0, 0.0])
        n = np.cross(t, ref); n /= np.linalg.norm(n) + 1e-9
        if prev_n is not None and np.dot(n, prev_n) < 0: n = -n
        prev_n = n
        b = np.cross(t, n)
        if i: vacc += np.linalg.norm(path[i] - path[i - 1])
        ring = []
        for q in range(sides + 1):
            a = 2 * np.pi * q / sides
            r = radii[i] * (1.0 + bumpy * (rng.random() - 0.5)) if q < sides else None
            if q == sides: ring.append(ring[0][0:1] + [ring[0][1]]); continue
            pos = p + r * (np.cos(a) * n + np.sin(a) * b)
            ring.append([add_vert(pos, (uscale * q / sides, vscale * vacc)), None])
        rings.append([x[0] for x in ring])
    # the seam vertex duplicates for the texture wrap
    for i, ring in enumerate(rings):
        first = verts[ring[0]]
        ring[-1] = add_vert(first[0], (uscale, first[1][1]))
    for i in range(len(rings) - 1):
        for q in range(sides):
            a0, a1, b0, b1 = rings[i][q], rings[i][q + 1], rings[i + 1][q], rings[i + 1][q + 1]
            faces.append((a0, a1, b1, tex)); faces.append((a0, b1, b0, tex))
    if cap_top:
        c = add_vert(path[-1] + 0.3 * radii[-1] * (path[-1] - path[-2]) / (np.linalg.norm(path[-1] - path[-2]) + 1e-9), (0.5, vscale * vacc))
        for q in range(sides):
            faces.append((rings[-1][q], rings[-1][q + 1], c, tex))

def blob(center, radius, tex, n=10, m=7, squash=1.0):
    """a lumpy ball (brain coral, rock lumps)"""
    center = np.asarray(center, float)
    ids = []
    for j in range(m + 1):
        th = np.pi * j / m
        row = []
        for i in range(n + 1):
            ph = 2 * np.pi * i / n
            r = radius * (1 + 0.15 * (rng.random() - 0.5))
            p = center + r * np.array([np.sin(th) * np.cos(ph), np.sin(th) * np.sin(ph), squash * np.cos(th)])
            row.append(add_vert(p, (i / n, j / m)))
        ids.append(row)
    for j in range(m):
        for i in range(n):
            a0, a1, b0, b1 = ids[j][i], ids[j][i + 1], ids[j + 1][i], ids[j + 1][i + 1]
            faces.append((a0, b0, b1, tex)); faces.append((a0, b1, a1, tex))

ROCK, CORAL = 0, 1

# the base: a low, rough mound of reef rock
base_path = [(0, 0, -1.5), (0, 0, 0.3), (0, 0, 1.2), (0, 0, 1.6)]
tube(base_path, [9.5, 9.0, 7.0, 3.5], 18, ROCK, uscale=4, vscale=0.3, bumpy=0.25)

def branch(p, d, length, radius, depth):
    """branching coral, like staghorn"""
    pts = [p]
    for k in range(1, 5):
        d = d + 0.15 * rng.standard_normal(3); d[2] = abs(d[2]) + 0.3; d /= np.linalg.norm(d)
        pts.append(pts[-1] + d * length / 4)
    tube(pts, [radius * (1 - 0.6 * k / 4) for k in range(5)], 6, CORAL, uscale=1, vscale=0.5)
    if depth > 0:
        for s in (-1, 1):
            nd = d + s * 0.6 * np.cross(d, [0, 0, 1]) + 0.2 * rng.standard_normal(3)
            branch(pts[-1], nd / np.linalg.norm(nd), length * 0.65, radius * 0.6, depth - 1)

hitboxes = [((-9.5, -9.5, -1.5), (9.5, 9.5, 1.6))]
for k in range(3):
    a = 2 * np.pi * k / 3 + 0.3
    c = np.array([5.2 * np.cos(a), 5.2 * np.sin(a), 0.0])
    h = 13.0 + 1.5 * k
    # a twisted column of reef rock
    path, radii = [], []
    for i in range(10):
        t = i / 9
        wob = 0.6 * np.array([np.sin(t * 5 + k), np.cos(t * 4 + 2 * k), 0])
        path.append(c + wob + np.array([0, 0, t * h]))
        radii.append(2.4 - 0.9 * t + 0.3 * np.sin(t * 9 + k))
    tube(path, radii, 12, ROCK, uscale=2, vscale=0.2, bumpy=0.3)
    top = path[-1]
    # lumps and brain corals on the column
    for i in range(4):
        t = rng.uniform(0.15, 0.85)
        ang = rng.uniform(0, 2 * np.pi)
        pi_ = int(t * 9)
        blob(path[pi_] + 1.6 * np.array([np.cos(ang), np.sin(ang), 0]), rng.uniform(0.8, 1.3), CORAL if i % 2 else ROCK, squash=0.8)
    # the coral crown
    for i in range(5):
        ang = 2 * np.pi * i / 5 + rng.uniform(-0.3, 0.3)
        d = np.array([0.55 * np.cos(ang), 0.55 * np.sin(ang), 1.0]); d /= np.linalg.norm(d)
        branch(top + np.array([0, 0, -0.5]), d, rng.uniform(2.6, 3.4), 0.5, 2)
    blob(top + np.array([0, 0, 0.6]), 1.4, CORAL, squash=0.7)
    hitboxes.append((tuple(c - [2.5, 2.5, 0]), tuple(c + [2.5, 2.5, h + 4])))

# normals: area-weighted per vertex
P = np.array([v[0] for v in verts]); UV = np.array([v[1] for v in verts])
N = np.zeros_like(P); FN = []
for a, b, c, t in faces:
    n = np.cross(P[b] - P[a], P[c] - P[a])
    N[a] += n; N[b] += n; N[c] += n
    FN.append(n / (np.linalg.norm(n) + 1e-9))
N /= np.linalg.norm(N, axis=1, keepdims=True) + 1e-9

with open(os.path.join(out, 'manalith.bin'), 'wb') as f:
    texnames = [b'HRCK', b'HCRL']  # the faces' texture names (files are found by these, see hydromod.d)
    f.write(b'HMNL')
    f.write(struct.pack('<IIII', len(verts), len(faces), len(texnames), len(hitboxes)))
    for t in texnames: f.write(t)
    for (lo, hi) in hitboxes: f.write(struct.pack('<6f', *lo, *hi))
    for i in range(len(verts)): f.write(struct.pack('<8f', *P[i], *UV[i], *N[i]))
    for (a, b, c, t), n in zip(faces, FN): f.write(struct.pack('<4I3f', a, b, c, t, *n))
print('vertices', len(verts), 'faces', len(faces), 'height', P[:, 2].max())

# textures
def noise(N_, cell):
    n = rng.random((N_ // cell, N_ // cell))
    return np.array(Image.fromarray((n * 255).astype(np.uint8)).resize((N_, N_), Image.BICUBIC)) / 255
S = 256
n1, n2, n3 = noise(S, 32), noise(S, 8), noise(S, 4)
pores = (n3 > 0.72).astype(float) * 0.5
rock = np.stack([0.72 + 0.15 * n1, 0.58 + 0.12 * n1, 0.50 + 0.1 * n1], -1) * (0.8 + 0.3 * n2)[..., None]  # sandy reef rock
rock = rock * (1 - 0.3 * (n2 > 0.65)[..., None]) + np.array([0.85, 0.45, 0.55]) * 0.3 * (n2 > 0.65)[..., None]  # pink coralline crust
rock *= (1 - 0.45 * pores)[..., None]
rock = rock * (1 - 0.25 * (n1 > 0.6)[..., None]) + np.array([0.35, 0.45, 0.3]) * 0.25 * (n1 > 0.6)[..., None]  # algae stains
Image.fromarray((np.clip(rock, 0, 1) * 255).astype(np.uint8)).save(os.path.join(out, 'reefrock.png'))
v, u = np.mgrid[0:S, 0:S] / (S - 1.0)
band = 0.5 + 0.5 * np.sin(2 * np.pi * (v * 3 + n1))
coral = np.stack([0.95 * np.ones_like(u), 0.35 + 0.25 * band, 0.25 + 0.2 * (1 - band)], -1)  # orange to coral red
coral = coral * (0.8 + 0.25 * n2)[..., None]
dots = (noise(S, 3) > 0.75)
coral[dots] = coral[dots] * 0.6 + np.array([1.0, 0.9, 0.8]) * 0.4  # polyps
Image.fromarray((np.clip(coral, 0, 1) * 255).astype(np.uint8)).save(os.path.join(out, 'coral.png'))
