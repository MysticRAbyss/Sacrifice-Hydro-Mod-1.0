#!/usr/bin/env python3
"""Build the spoken lines of the talking Hydro creatures with the Piper speech synthesiser.

Voice: Piper "en-us-libritts-high" (trained on LibriTTS, CC BY 4.0, http://www.openslr.org/60/),
one of its 904 speakers per creature. Piper: https://github.com/rhasspy/piper (MIT).

usage: make_lines.py <piper binary dir> <libritts .onnx model> <out dir>
Needs ffmpeg (with rubberband) and numpy. Output: 22050 Hz 16-bit mono WAV,
<out>/<creature>/line_<type><n>.wav, plus <out>/lines.json listing them per creature and type.
"""
import json, os, subprocess, sys, tempfile, wave
import numpy as np

SR = 22050

# speaker: LibriTTS speaker index; speed: Piper length scale (>1 slower); pitch: semitone factor
# (rubberband, keeps the speed); fx: extra ffmpeg filters after the pitch shift
CREATURES = {
 'tidecaller': dict(speaker=186, speed=1.15, pitch=0.85,
    fx="lowpass=f=5000,chorus=0.6:0.8:40|60:0.3|0.25:0.3|0.4:2|1.5,aecho=0.8:0.6:60|120:0.35|0.2",
    lines=dict(
     selected=["The tide listens.", "Speak, and the deep will answer.", "Salt and silence."],
     moving=["As the current wills.", "I drift where I am needed."],
     attacking=["Drown in your own breath!", "The sea takes what it is owed!"],
     annoyed=["Patience. Even the ocean waits for the moon."],
     damaged=["The water... grows cold."],
     death=["Return me... to the deep..."])),
 'leviathan': dict(speaker=137, speed=1.2, pitch=0.68,
    fx="lowpass=f=3500,chorus=0.5:0.9:50|70|90:0.4|0.35|0.3:0.25|0.3|0.35:2|2.3|1.7,aecho=0.8:0.7:90|180|300:0.4|0.25|0.15",
    lines=dict(
     selected=["Who dares wake me?", "I was old when your mountains were young."],
     moving=["The world is a puddle.", "Make way."],
     attacking=["Feel the weight of the abyss!", "Swallowed whole!"],
     annoyed=["Poke me again, little wizard. I dare you."],
     damaged=["You only scratch the scales."],
     death=["The sea... remembers..."])),
 'abyssguardian': dict(speaker=49, speed=1.25, pitch=0.62,
    fx="lowpass=f=2800,acrusher=bits=10:mix=0.25:samples=1,aecho=0.8:0.6:50|110:0.35|0.2",
    lines=dict(
     selected=["Guard.", "Deep... watches."],
     moving=["Go."],
     attacking=["Crush!", "Down!"],
     annoyed=["Hmph."],
     death=["Sink..."])),
 'reefstalker': dict(speaker=42, speed=1.05, pitch=0.92,
    fx="highpass=f=120,lowpass=f=6000,aecho=0.8:0.5:35|70:0.2|0.1",
    lines=dict(
     selected=["I see them before they see me.", "The reef hides many teeth."],
     moving=["Silent as the current.", "Following the scent."],
     attacking=["One shot.", "Hold still, little fish."],
     annoyed=["You'll scare the prey."],
     damaged=["Just a scratch... of coral."],
     death=["The hunt... ends..."])),
 'tidekin': dict(speaker=350, speed=0.92, pitch=1.22,
    fx="highpass=f=150,lowpass=f=7000,aecho=0.8:0.5:30|60:0.2|0.1",
    lines=dict(
     selected=["Tidekin, reporting!", "Shell's polished, ready to go!"],
     moving=["Splish splash!", "On my way!"],
     attacking=["Pinch 'em!", "For the tide!"],
     annoyed=["Stop tapping my shell, it echoes in here!"],
     damaged=["Ow, my barnacles!"],
     death=["Tell my pod... I was brave..."])),
}

def speak(piper_dir, model, text, c, path):
    env = dict(os.environ, LD_LIBRARY_PATH=piper_dir)
    with tempfile.NamedTemporaryFile(suffix='.wav') as raw, tempfile.NamedTemporaryFile(suffix='.wav') as fx:
        subprocess.run([os.path.join(piper_dir, 'piper'), '-m', model, '-s', str(c['speaker']),
                        '--length_scale', str(c['speed']), '--noise_scale', '0.5', '--noise_w', '0.6',
                        '--sentence_silence', '0', '-f', raw.name],
                       input=text.encode(), check=True, env=env, capture_output=True)
        af = f"rubberband=pitch={c['pitch']}:formant=shifted,{c['fx']}"
        subprocess.run(['ffmpeg', '-loglevel', 'error', '-y', '-i', raw.name, '-af', af, '-ac', '1', '-ar', str(SR),
                        '-sample_fmt', 's16', fx.name], check=True)
        with wave.open(fx.name) as w: x = np.frombuffer(w.readframes(w.getnframes()), np.int16).astype(float)
    a = np.abs(x); idx = np.nonzero(a > a.max() * 0.02)[0]
    x = x[max(0, idx[0] - 100):idx[-1] + 1 + int(0.25 * SR)]
    n = min(len(x) // 4, int(0.05 * SR)); x[-n:] *= np.linspace(1, 0, n)
    x = x / np.abs(x).max() * 32767 * 0.8
    with wave.open(path, 'wb') as w:
        w.setnchannels(1); w.setsampwidth(2); w.setframerate(SR); w.writeframes(x.astype(np.int16).tobytes())

def main(piper_dir, model, out):
    manifest = {}
    for name, c in CREATURES.items():
        os.makedirs(os.path.join(out, name), exist_ok=True)
        js = {}
        for typ, texts in c['lines'].items():
            js[typ] = []
            for i, t in enumerate(texts, 1):
                fn = f"line_{typ}{i}.wav"
                speak(piper_dir, model, t, c, os.path.join(out, name, fn))
                js[typ].append(f"{name}/{fn}")
        manifest[name] = js
    with open(os.path.join(out, 'lines.json'), 'w') as f: json.dump(manifest, f, indent=1)

if __name__ == '__main__':
    main(*sys.argv[1:4])
