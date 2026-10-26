#!/usr/bin/env python3
"""Build the Hydro creature voices from three CC0 sound packs.

Packs (CC0, opengameart.org):
  a = "80 CC0 creature SFX"   by rubberduck  (80-CC0-creature-SFX_0.zip)
  b = "80 CC0 creature SFX 2" by rubberduck  (80-CC0-creature-sfx-2.zip)
  m = "Monster SFX pack"      by Ogrebane    (monster_sfx_pack.zip)

usage: make_voices.py <dir with the three zips unpacked> <out dir>
Needs ffmpeg and numpy. Output: 22050 Hz 16-bit mono WAV, <out>/<creature>/<type><n>.wav
"""
import json, os, subprocess, sys, tempfile, wave
import numpy as np

SR = 22050

def src(packs, ref):
    p, n = ref.split('_', 1)
    if p == 'a': return os.path.join(packs, '80-CC0-creature-SFX_0', n + '.ogg')
    if p == 'b': return os.path.join(packs, '80-CC0-creature-sfx-2', n + '.ogg')
    return os.path.join(packs, 'monster_sfx_pack', 'monster_sfx_pack', n + '.wav')

# per creature: pitch factor (<1 deeper/slower), lowpass Hz, echo (watery room), clips per sound type
CREATURES = {
 'tidekin': dict(pitch=1.0, lp=6000, echo=0.20, sounds=dict(
    selected=['a_burble_01', 'a_cute_04'], annoyed=['a_burp_02'], moving=['a_cute_05', 'a_cute_09'],
    attacking=['a_grunt_04', 'b_attack_01'], damaged=['a_hurt_01', 'a_hurt_05'], death=['b_die_01'])),
 'reefstalker': dict(pitch=0.85, lp=5000, echo=0.20, sounds=dict(
    selected=['b_human_05', 'b_human_03'], annoyed=['b_human_01'], moving=['b_human_06', 'a_spit_03'],
    attacking=['b_attack_02', 'a_spit_02'], damaged=['b_human_02', 'a_cough_02'], death=['b_die_03'])),
 'mistwing': dict(pitch=1.1, lp=7000, echo=0.25, sounds=dict(
    selected=['a_alien_02', 'a_alien_03'], annoyed=['a_scream_02'], moving=['a_weird_02', 'b_alien_12'],
    attacking=['a_alien_04', 'b_alien_08'], damaged=['a_hurt_02', 'b_bug_10'], death=['b_die_04'])),
 'mantisshrimp': dict(pitch=0.9, lp=7000, echo=0.15, sounds=dict(
    selected=['b_bug_06', 'b_bug_08'], annoyed=['b_bug_11'], moving=['a_bug_01', 'a_bug_04'],
    attacking=['b_bug_13', 'a_eat_04'], damaged=['b_bug_05', 'a_bug_02'], death=['b_bug_07'])),
 'brineclaw': dict(pitch=0.8, lp=5000, echo=0.20, sounds=dict(
    selected=['a_misc_02', 'b_misc_15'], annoyed=['a_eat_02'], moving=['a_misc_08', 'b_misc_12'],
    attacking=['b_monster_11', 'b_monster_10'], damaged=['a_hurt_03', 'b_misc_19'], death=['b_die_02'])),
 'tidecaller': dict(pitch=0.75, lp=4500, echo=0.30, sounds=dict(
    selected=['b_human_04', 'a_ooh'], annoyed=['b_human_07'], moving=['b_human_08', 'a_misc_09'],
    attacking=['a_howl', 'a_troll_03'], damaged=['a_grunt_03', 'b_monster_15'], death=['b_weird_06'])),
 'puffy': dict(pitch=1.15, lp=6000, echo=0.20, sounds=dict(
    selected=['a_burble_02', 'a_cute_01'], annoyed=['a_burp_01'], moving=['b_slime_07', 'a_cute_08'],
    attacking=['a_spit_01', 'b_slime_02'], damaged=['b_slime_04', 'a_cough_01'], death=['b_slime_08'])),
 'reefbreaker': dict(pitch=0.85, lp=3500, echo=0.25, sounds=dict(
    selected=['b_monster_09', 'b_monster_12'], annoyed=['a_monster_06'], moving=['b_grunt_06', 'b_misc_11'],
    attacking=['b_misc_10', 'b_monster_08'], damaged=['a_hurt_04', 'b_weird_09'], death=['b_monster_20'])),
 'abyssguardian': dict(pitch=0.8, lp=2500, echo=0.30, sounds=dict(
    selected=['m_monster-1', 'm_monster-4'], annoyed=['m_monster-6'], moving=['b_grunt_07', 'b_grunt_09'],
    attacking=['m_monster-2', 'm_monster-3', 'm_monster-7'], damaged=['m_monster-9', 'b_grunt_08'], death=['m_monster-8'])),
 'leviathan': dict(pitch=0.7, lp=2500, echo=0.35, sounds=dict(
    selected=['b_roar_04', 'a_roar_02'], annoyed=['b_roar_06'], moving=['a_monster_07', 'a_monster_04'],
    attacking=['b_roar_05', 'a_monster_03'], damaged=['m_monster-5', 'a_weird_05'], death=['m_monster-10'])),
}

def process(path, pitch, lp, echo):
    # resampling changes pitch and speed together (like a tape), which keeps the voice natural
    af = f"aresample={SR},asetrate={int(SR*pitch)},aresample={SR},highpass=f=60,lowpass=f={lp}"
    if echo > 0: af += f",aecho=0.8:0.7:35|70:{echo}|{echo/2}"
    with tempfile.NamedTemporaryFile(suffix='.wav') as t:
        subprocess.run(['ffmpeg', '-loglevel', 'error', '-y', '-i', path, '-af', af, '-ac', '1', '-ar', str(SR),
                        '-sample_fmt', 's16', t.name], check=True)
        with wave.open(t.name) as w: x = np.frombuffer(w.readframes(w.getnframes()), np.int16).astype(float)
    # trim leading/trailing silence, short fades, normalise to -2 dBFS
    a = np.abs(x); thr = a.max() * 0.02; idx = np.nonzero(a > thr)[0]
    x = x[max(0, idx[0] - 100):idx[-1] + 1 + int(0.15 * SR)]
    n = min(len(x) // 4, int(0.03 * SR)); x[-n:] *= np.linspace(1, 0, n); m = min(len(x) // 4, 40); x[:m] *= np.linspace(0, 1, m)
    x = x / np.abs(x).max() * 32767 * 0.8
    return x.astype(np.int16)

def main(packs, out):
    manifest = {}
    for name, c in CREATURES.items():
        js = {}
        os.makedirs(os.path.join(out, name), exist_ok=True)
        for typ, refs in c['sounds'].items():
            files = []
            for i, r in enumerate(refs, 1):
                y = process(src(packs, r), c['pitch'], c['lp'], c['echo'])
                fn = f"{typ}{i}.wav"
                with wave.open(os.path.join(out, name, fn), 'wb') as w:
                    w.setnchannels(1); w.setsampwidth(2); w.setframerate(SR); w.writeframes(y.tobytes())
                files.append(f"{name}/{fn}")
            js[typ] = files
        js['melee'] = js['attacking']; js['rescued'] = js['selected']
        manifest[name] = js
    with open(os.path.join(out, 'manifest.json'), 'w') as f: json.dump(manifest, f, indent=1)

if __name__ == '__main__':
    main(sys.argv[1], sys.argv[2])
