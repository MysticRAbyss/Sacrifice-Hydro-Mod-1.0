# Synthesised water sounds for the Hydro creatures' abilities (22050 Hz, mono, 16-bit WAV).
# Everything here is generated from noise and sine waves, so the sounds are free to share.
import numpy as np, wave, sys, os
SR=22050
rng=np.random.default_rng(1234)
def t(d): return np.arange(int(SR*d))/SR
def env(n,a,d):  # attack/decay envelope in seconds
    x=np.arange(n)/SR
    return np.minimum(1,x/max(a,1e-4))*np.exp(-x/max(d,1e-4))
def bandpass(x,lo,hi):
    X=np.fft.rfft(x); f=np.fft.rfftfreq(len(x),1/SR)
    X*=np.exp(-((np.log(np.maximum(f,1))-np.log(np.sqrt(lo*hi)))**2)/(2*(np.log(hi/lo)/2)**2))
    return np.fft.irfft(X,len(x))
def lowpass(x,fc):
    X=np.fft.rfft(x); f=np.fft.rfftfreq(len(x),1/SR); X/=np.sqrt(1+(f/fc)**4); return np.fft.irfft(X,len(x))
def bubble(f0,dur=0.06,rise=1.8):
    tt=t(dur); f=f0*(1+(rise-1)*tt/dur)
    return np.sin(2*np.pi*np.cumsum(f)/SR)*np.exp(-tt/(dur/3))*np.minimum(1,tt/0.003)
def place(out,s,at):
    i=int(at*SR); j=min(len(out),i+len(s)); out[i:j]+=s[:j-i]
def bubbles(out,start,end,rate,flo,fhi,gain=0.25):
    for _ in range(int((end-start)*rate)):
        place(out,gain*rng.uniform(0.4,1)*bubble(rng.uniform(flo,fhi),rng.uniform(0.03,0.09)),rng.uniform(start,end))
def splash(dur,lo,hi,decay):
    n=int(SR*dur); x=bandpass(rng.standard_normal(n),lo,hi); return x/np.abs(x).max()*env(n,0.004,decay)
def droplets(out,start,end,count,gain=0.2):
    for _ in range(count):
        place(out,gain*bubble(rng.uniform(1200,3200),rng.uniform(0.01,0.025),rng.uniform(1.2,2.5)),rng.uniform(start,end))
def save(name,x):
    x=x/np.abs(x).max()*0.85
    fade=int(0.01*SR); x[-fade:]*=np.linspace(1,0,fade)
    with wave.open(name,'wb') as w:
        w.setnchannels(1); w.setsampwidth(2); w.setframerate(SR)
        w.writeframes((x*32767).astype('<i2').tobytes())

out_dir=sys.argv[1] if len(sys.argv)>1 else '.'
os.makedirs(out_dir,exist_ok=True)
P=lambda n: os.path.join(out_dir,n)

# Maelstrom: deep churning water with a slow swirl, gurgles and bubbles (3.5 s)
d=3.5; n=int(SR*d); tt=t(d)
roar=lowpass(rng.standard_normal(n),420)*(0.6+0.4*np.sin(2*np.pi*0.9*tt))
swirl=np.zeros(n)
for k in range(40):  # sweeping bandpassed noise grains = the swirl
    g=int(0.18*SR); s=bandpass(rng.standard_normal(g),300+600*(k%5),900+1400*(k%5))*np.hanning(g)
    place(swirl,s,k*d/40)
x=roar/np.abs(roar).max()+0.5*swirl/np.abs(swirl).max()
bubbles(x,0.1,d-0.2,40,120,500,0.35)
x*=np.minimum(1,tt/0.35)*np.minimum(1,(d-tt)/0.8)
save(P('maelstrom.wav'),x)

# Whirl: a short whirlpool burst (Reefbreaker's mortar) (1.4 s)
d=1.4; n=int(SR*d); tt=t(d)
x=splash(d,250,2500,0.25)+0.6*lowpass(rng.standard_normal(n),500)*np.exp(-tt/0.6)
bubbles(x,0.05,1.2,45,150,700,0.4); droplets(x,0.02,0.6,25)
save(P('whirl.wav'),x)

# Tidal Slam: a heavy body hitting water, a big splash and falling spray (1.8 s)
d=1.8; n=int(SR*d); tt=t(d)
thud=np.sin(2*np.pi*np.cumsum(70*np.exp(-tt/0.3)+40)/SR)*env(n,0.002,0.18)
x=1.2*thud+splash(d,200,3000,0.35)+0.5*splash(d,1500,6000,0.12)
droplets(x,0.15,1.5,60,0.18); bubbles(x,0.1,1.0,30,100,400,0.3)
save(P('slam.wav'),x)

# Tidal Mend: soft bubbling with a gentle watery shimmer (1.6 s)
d=1.6; n=int(SR*d); tt=t(d); x=np.zeros(n)
for i,f in enumerate([523.25,659.25,783.99,1046.5,1318.5]):
    x+=0.18*np.sin(2*np.pi*f*tt+3*np.sin(2*np.pi*5*tt))*env(n,0.05+0.08*i,0.7)
bubbles(x,0.0,1.3,30,400,1100,0.25)
save(P('mend.wav'),x)

# Shell Up: a hard shell clamping shut with a bubbly gulp (0.8 s)
d=0.8; n=int(SR*d); tt=t(d)
x=np.sin(2*np.pi*np.cumsum(180*np.exp(-tt/0.08)+90)/SR)*env(n,0.001,0.1)
x+=0.7*splash(d,2000,7000,0.025)
bubbles(x,0.05,0.5,25,180,500,0.4)
save(P('shell.wav'),x)

# Toxic Spines: a burst of spines whooshing out, with a wet spit (0.9 s)
d=0.9; n=int(SR*d); tt=t(d); x=np.zeros(n)
for k in range(8):
    g=int(0.25*SR); gt=np.arange(g)/SR; s=rng.standard_normal(g)
    s=bandpass(s,1500,7000)*np.exp(-gt/0.06)*np.minimum(1,gt/0.004)
    place(x,s/np.abs(s).max()*rng.uniform(0.5,1),0.02+k*0.03)
x+=0.6*splash(d,400,2500,0.08)
save(P('spines.wav'),x)

# Splash hit: a small water bolt hitting (Mantis Shrimp / Puffy) (0.5 s)
d=0.5; n=int(SR*d); x=splash(d,500,4000,0.07); droplets(x,0.02,0.3,12,0.25)
save(P('splash.wav'),x)
# Water Jet: a pressurised rush of water leaving the hands (0.9 s)
d=0.9; n=int(SR*d); tt=t(d)
x=bandpass(rng.standard_normal(n),600,5000)*np.minimum(1,tt/0.03)*np.exp(-np.maximum(0,tt-0.25)/0.25)
x=x/np.abs(x).max()
x+=0.4*np.sin(2*np.pi*np.cumsum(300+500*tt)/SR)*env(n,0.02,0.3)
droplets(x,0.1,0.8,25,0.25)
save(P('jet.wav'),x)

# Bubble Armor: a big bubble swelling around the wizard, with a soft wobble (1.4 s)
d=1.4; n=int(SR*d); tt=t(d)
x=np.sin(2*np.pi*np.cumsum(140+260*(1-np.exp(-tt/0.35))+18*np.sin(2*np.pi*7*tt))/SR)*env(n,0.08,0.6)
x+=0.35*lowpass(rng.standard_normal(n),900)*env(n,0.05,0.4)
bubbles(x,0.0,1.1,40,250,900,0.35)
save(P('bubble.wav'),x)
# Geyser: a deep rumble, then a roaring blast of water that rains back down (3 s)
d=3.0; n=int(SR*d); tt=t(d); x=np.zeros(n)
rum=lowpass(rng.standard_normal(n),120)*np.clip(tt/0.9,0,1)*(tt<0.95)
x+=rum/np.abs(rum).max()*0.6
blast=bandpass(rng.standard_normal(n),200,3000)*np.where(tt>0.9,np.exp(-(tt-0.9)/0.7),0)
x+=blast/np.abs(blast).max()
droplets(x,1.2,2.9,120,0.2)
save(P('geyser.wav'),x)

# Healing Spring: a bright bubbling spring with a soft chime (2.2 s)
d=2.2; n=int(SR*d); tt=t(d); x=np.zeros(n)
for f0,a in ((880,0.35),(1320,0.25),(1760,0.15)):
    x+=a*np.sin(2*np.pi*f0*tt)*env(n,0.01,0.9)
x+=0.3*bandpass(rng.standard_normal(n),800,4000)*env(n,0.2,1.0)
bubbles(x,0.1,2.0,45,500,1400,0.3)
save(P('spring.wav'),x)

# Tidal Wall: a rushing curtain of water, made to loop (3 s)
d=3.0; n=int(SR*d); tt=t(d)
x=bandpass(rng.standard_normal(n),300,5000)
x=x/np.abs(x).max()*(0.8+0.2*np.sin(2*np.pi*tt/d*3))
droplets(x,0.0,2.95,80,0.15)
f=int(0.05*SR); x[:f]*=np.linspace(0,1,f)  # soft ends so the loop does not click
save(P('tide.wav'),x)
# Acid Rain: steady hissing rain with sizzles, made to loop (3 s)
d=3.0; n=int(SR*d); tt=t(d)
x=bandpass(rng.standard_normal(n),2000,9000); x=x/np.abs(x).max()*0.6
for _ in range(90):
    g=int(rng.uniform(0.03,0.12)*SR); s_=bandpass(rng.standard_normal(g),3000,9000)*np.exp(-np.arange(g)/SR/0.03)
    place(x,0.5*s_/np.abs(s_).max(),rng.uniform(0,d-0.15))
f=int(0.05*SR); x[:f]*=np.linspace(0,1,f)
save(P('acid.wav'),x)

# Tsunami: a deep, building roar of a wave crashing (4 s)
d=4.0; n=int(SR*d); tt=t(d)
x=lowpass(rng.standard_normal(n),400)*np.clip(tt/1.5,0,1)**2*np.exp(-np.maximum(0,tt-2.5)/0.8)
x=x/np.abs(x).max()
x+=0.6*bandpass(rng.standard_normal(n),500,5000)*np.clip((tt-1.2)/0.6,0,1)*np.exp(-np.maximum(0,tt-1.8)/0.9)
droplets(x,1.8,3.8,150,0.15)
save(P('tsunami.wav'),x)

# Kraken: a churning whirlpool under a low, groaning bellow (4 s)
d=4.0; n=int(SR*d); tt=t(d)
x=0.7*lowpass(rng.standard_normal(n),200)*env(n,0.4,2.5)
f0=55+8*np.sin(2*np.pi*0.7*tt)
x+=0.8*np.sin(2*np.pi*np.cumsum(f0)/SR)*np.sin(np.pi*np.clip(tt/2.6,0,1))*(1+0.3*np.sin(2*np.pi*13*tt))
bubbles(x,0.2,3.6,40,120,400,0.35)
save(P('kraken.wav'),x)
print('ok')
