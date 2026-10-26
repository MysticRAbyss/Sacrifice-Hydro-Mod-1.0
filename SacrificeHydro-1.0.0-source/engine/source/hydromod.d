// Hydro mod for sacengine: loads new creatures (models, textures, icons, names, stats) from
// mods/hydro/creatures/<name>/creature.json next to the executable.
// distributed under the terms of the gplv3 license
// https://www.gnu.org/licenses/gpl-3.0.txt
module hydromod;

import std.algorithm, std.conv, std.exception, std.file, std.json, std.path, std.stdio, std.string;
import util, spells, nttData, assets;

string modsRoot; // folder that contains mods/ (the executable's folder)
string[] hydroLaunchArgs; // the command line the game was started with (to restart it with a saved game)
bool[string] rawUVModels; // models whose UVs are stored directly (any texture size)
int[][string] noDispModels;
string[string] hydroAnimOwner; // animation copies made for a new creature -> that creature's model file // model file -> bones that stay on the ground (ignore the root displacement)
// boneMap: base bone -> our bone; rebind: base bone -> base bone whose (rest pose) frame its
// vertices are re-expressed in first (e.g. a wing's root vertices that hang off the chest)
struct BorrowedPart{ int part; int[int] boneMap; float scale=1.0f; int[int] rebind; }
struct BorrowedParts{ string baseModel; BorrowedPart[] parts; }
BorrowedParts[string] borrowedParts; // model file -> body parts to take from the base creature's model
struct HydroRangedStats{ float amount=float.nan, range=float.nan; }
HydroRangedStats[char[4]] hydroRangedStats;
bool[char[4]] hydroStartsFlying; // creatures that are already flying when they are summoned ("startsFlying")
bool[char[4]] hydroBoltCreatures; // creatures whose ranged attack is drawn as a water bolt instead of an arrow ("boltProjectile")
int[char[4]] hydroShootBone; // creatures that shoot from the top of the part on this bone (e.g. a staff) ("shootFrom") // creature tag -> its own damage and range for its ranged attack (creature.json "rangedAttack")
string[char[4]][char[4]] hydroAbilityNames; // creature tag -> ability tag -> name shown for that creature's ability
enum char[4] maelstromTag="tslm"; // the Leviathan's Maelstrom ability (reads "mlst")
enum char[4] tidalMendTag="dmdt"; // the Tidecaller's Tidal Mend ability (reads "tdmd")
enum char[4] shellUpTag="puhs"; // the Tidekin's Shell Up ability (reads "shup")
enum char[4] toxicSpinesTag="npst"; // Puffy's Toxic Spines ability (reads "tspn")
enum char[4] mantisTag="rhsm", reefbreakerTag="rbfr", abyssGuardianTag="dgba", leviathanTag="ntvl";

// the Maelstrom: a smaller, shorter whirlwind built from the Tornado spell's data, cast instantly
// around the caster
private Spel synthetic(char[4] baseTag,char[4] tag,string name,uint manaCost,uint cooldownSeconds){
	auto base=baseTag in spels;
	enforce(base,text("hydro mod: no spell ",baseTag));
	Spel m=*base;
	m.name=tag;
	m.manaCost=manaCost;
	m.cooldown=cooldownSeconds*60;
	m.flags=cast(SpelFlags)0; // no target: cast around the caster
	m.castingTime=0;
	m.minCastingTime=0;
	m.spellOrder=0;
	(cast(string[char[4]])texts)[tag]=name;
	return m;
}
private void addMaelstrom(){ // the synthetic abilities of Hydro creatures
	if(maelstromTag in spels) return;
	auto spells=cast(Spel[char[4]])spels;
	// Maelstrom: a large whirlpool on the ground around the Leviathan (from the Vortick's vortex)
	auto m=synthetic(SpellTag.vortickShoot,maelstromTag,"Maelstrom",250,30);
	m.effectRange=12.0f; // radius of the whirlpool
	spells[maelstromTag]=m;
	// Tidal Mend: the Heal spell on every friendly creature around the caster
	auto t=synthetic(SpellTag.heal,tidalMendTag,"Tidal Mend",80,20);
	t.effectRange=20.0f;
	spells[tidalMendTag]=t;
	// Shell Up: the Tidekin hardens its shell for a while
	auto u=synthetic(SpellTag.rockForm,shellUpTag,"Shell Up",40,25);
	u.duration=12.0f;
	spells[shellUpTag]=u;
	// Toxic Spines: Puffy shoots poisoned spines all around it (harmless darts; the poison does the damage)
	auto p=synthetic(SpellTag.rangerShoot,toxicSpinesTag,"Toxic Spines",60,15);
	p.amount=0;
	p.effectRange=20.0f; // "200" in the stat scale: sight range 1000 is 100 meters
	p.duration=8.0f;
	spells[toxicSpinesTag]=p;
}

// Hydro's god spells: each starts from the matching spell of another god (its targeting, casting
// time, level and mana cost) and gets its own name, icon and behaviour (state.d, "Hydro mod spell")
private Spel godSpell(char[4] baseTag,char[4] tag,string name){
	auto base=baseTag in spels;
	enforce(base,text("hydro mod: no spell ",baseTag));
	Spel m=*base;
	m.name=tag;
	(cast(string[char[4]])texts)[tag]=name;
	return m;
}
private void addHydroSpells(string root){
	if(SpellTag.waterJet in spels) return;
	auto spells=cast(Spel[char[4]])spels;
	// Water Jet: the Fireball's slot; a jet of water that hits and knocks back
	auto w=godSpell(SpellTag.fireball,SpellTag.waterJet,"Water Jet");
	spells[SpellTag.waterJet]=w;
	// Bubble Armor: the Air Shield's slot; -60% damage taken and slow regeneration for 40 s, 60 s cooldown
	auto b=godSpell(SpellTag.airShield,SpellTag.bubbleArmor,"Bubble Armor");
	b.duration=40.0f; // seconds
	b.cooldown=60*60; // 60 seconds (in frames)
	spells[SpellTag.bubbleArmor]=b;
	// Undertow: the Freeze's slot; a whirling pool around the target that holds enemies, then slows them
	auto u=godSpell(SpellTag.freeze,SpellTag.undertow,"Undertow");
	u.effectRange=10.0f; // radius of the pool
	u.duration=4.0f; // how long the pool holds the target and slows the others
	spells[SpellTag.undertow]=u;
	// Healing Spring: level 4 (mana and casting like Erupt, the level-4 spell of James); cast on a
	// friendly creature like the Healing Aura, a spring opens at its feet and heals everyone of the
	// caster's side standing in it
	auto h=godSpell(SpellTag.healingAura,SpellTag.healingSpring,"Healing Spring");
	if(auto e=SpellTag.erupt in spels){ h.manaCost=e.manaCost; h.cooldown=e.cooldown; h.castingTime=e.castingTime; h.minCastingTime=e.minCastingTime; }
	h.effectRange=24.0f; // radius of the spring
	h.duration=12.0f; // how long it flows
	h.amount=60; // health per second
	spells[SpellTag.healingSpring]=h;
	// Geyser: level 5, aimed like the Explosion; the ground bursts with a column of water that
	// damages and throws up the creatures around
	auto g=godSpell(SpellTag.explosion,SpellTag.geyser,"Geyser");
	g.effectRange=12.0f; // radius of the dome: every enemy on it is damaged and thrown
	spells[SpellTag.geyser]=g;
	// Tidal Wall: level 6, raised like the Firewall; a wall of falling water over a reef that blocks
	// enemy missiles, slows and lightly damages enemies walking through, and lets its side pass
	auto tw=godSpell(SpellTag.firewall,SpellTag.tidalWall,"Tidal Wall");
	spells[SpellTag.tidalWall]=tw;
	// Acid Rain: level 7, cast like the Rain of Fire; a dark cloud pours corrosive rain on a 20 m
	// radius for 12 s, eating 140 health per second from every enemy beneath it
	auto ar=godSpell(SpellTag.rainOfFire,SpellTag.acidRain,"Acid Rain");
	ar.effectRange=20.0f; ar.duration=12.0f;
	spells[SpellTag.acidRain]=ar;
	// Tsunami: level 8, cast like the Cloudkill; a wave 40 m wide rolls 75 m in the direction cast,
	// hitting every enemy once (medium-heavy damage) and sweeping it along, small ones further
	auto ts=godSpell(SpellTag.cloudkill,SpellTag.tsunami,"Tsunami");
	ts.amount=450; ts.duration=6.0f; ts.cooldown=60*60;
	spells[SpellTag.tsunami]=ts;
	// Kraken: level 8 (the last spell), cast like the Tornado; a whirlpool of 80 m radius drags
	// everyone (both sides) to its eye for 14 s, where four tentacles, 75% the size of Persephone's
	// Meanstalks, each grab a creature (any side, wizards too), crush it (350) and fling it
	auto kr=godSpell(SpellTag.tornado,SpellTag.kraken,"Kraken");
	kr.effectRange=80.0f; kr.duration=14.0f; kr.amount=350; // 350: the damage of each tentacle's grab
	if(auto m=SpellTag.meanstalks in spels){ kr.speed=m.speed; kr.acceleration=m.acceleration; } // tentacles reach like the Meanstalks' vines
	spells[SpellTag.kraken]=kr;
	{ import std.process:environment; if(environment.get("HYDRO_DEBUG","").length) foreach(tg;[SpellTag.healingAura,SpellTag.erupt,SpellTag.explosion,SpellTag.firewall,SpellTag.etherealForm,SpellTag.fireform,SpellTag.skinOfStone,SpellTag.airShield,SpellTag.protectiveSwarm,SpellTag.bubbleArmor,SpellTag.rainOfFire,SpellTag.cloudkill,SpellTag.tornado,SpellTag.volcano,SpellTag.blindRage,SpellTag.charm]) if(auto q=tg in spels) writeln("hydro mod: base ",tg," mana ",q.manaCost," amount ",q.amount," range ",q.range," effectRange ",q.effectRange," damageRange ",q.damageRange," duration ",q.duration," casting ",q.castingTime," cooldown ",q.cooldown," flags ",q.flags," flags1 ",q.flags1," flags2 ",q.flags2," speed ",q.speed); }
	// spellbar icons: mods/hydro/spells/<readable tag>.png (64x64)
	foreach(tag;[SpellTag.waterJet,SpellTag.bubbleArmor,SpellTag.undertow,SpellTag.healingSpring,SpellTag.geyser,SpellTag.tidalWall,SpellTag.acidRain,SpellTag.tsunami,SpellTag.kraken]){
		char[4] t=tag;
		string readable=[t[3],t[2],t[1],t[0]];
		auto png=buildPath(root,"spells",readable~".png");
		if(!exists(png)) continue;
		import dlib.image.io.png:loadPNG;
		textureReplacementsSloppy[t]=loadPNG(png); // found by the icon's file name below
		icons[t]=buildPath("extracted","hydro","spells",readable~".ICON");
		spells[t].icon=t;
	}
}

// Hydro's manalith: three pillars of reef rock crowned with coral (mods/hydro/manalith, made by
// tools/sfx/make_manalith.py). The mesh is put into the game's model format (MRMM) using the
// Stratos manalith's model as the template, and the building copies the Stratos manalith's.
private void addHydroManalith(string dir){
	import bldg:Bldg,BldgHeader,BldgComponent;
	import mrmm:parseMRMM;
	enum char[4] stratos="mats";
	auto binPath=buildPath(dir,"manalith.bin");
	if(!exists(binPath)||hydroManalithTag in bldgs||stratos!in bldgs||stratos!in bldgModls) return;
	try{
		auto bin=cast(ubyte[])read(binPath);
		enforce(bin[0..4]==cast(ubyte[])"HMNL","not a manalith mesh");
		uint u32(size_t o){ return *cast(uint*)&bin[o]; }
		auto nv=u32(4), nf=u32(8), nt=u32(12), nh=u32(16);
		size_t o=20;
		string[] texNames;
		foreach(i;0..nt){ texNames~=cast(string)bin[o..o+4].idup; o+=4; }
		auto hitboxBytes=bin[o..o+24*nh]; o+=24*nh;
		auto vertexData=bin[o..o+32*nv]; o+=32*nv;
		auto faceData=bin[o..o+28*nf];
		// the model file: the Stratos manalith's header, our counts, hitboxes, vertices and faces
		auto tmpl=readFile(bldgModls[stratos]);
		auto header=tmpl[0..192].dup;
		enum hitboxSize=28, vertexSize=40, faceSize=36;
		auto vertexOff=192+hitboxSize*nh, faceOff=vertexOff+vertexSize*nv, size=faceOff+faceSize*nf;
		ubyte[] data=header~new ubyte[](size-192);
		void put32(size_t at,uint v){ *cast(uint*)&data[at]=v; }
		void putf(size_t at,float v){ *cast(float*)&data[at]=v; }
		put32(0,cast(uint)size); put32(4,nv); put32(12,nf); put32(24,1); put32(68,nh);
		put32(184,cast(uint)vertexOff); put32(188,cast(uint)faceOff);
		foreach(i;0..nh){ // hitboxes live in the header area at 72..
			enforce(72+hitboxSize*(i+1)<=184,"too many hitboxes");
		}
		// (the header keeps room for 4 hitboxes at 72; ours go there)
		foreach(i;0..nh){
			auto at=72+hitboxSize*i;
			put32(at,0);
			foreach(k;0..6) putf(at+4+4*k,*cast(float*)&hitboxBytes[24*i+4*k]);
		}
		foreach(i;0..nv){
			auto src=vertexData[32*i..32*(i+1)], at=vertexOff+vertexSize*i;
			foreach(k;0..5) putf(at+4*k,*cast(float*)&src[4*k]); // position, uv
			put32(at+20,0xffffffff); put32(at+24,0);
			foreach(k;0..3) putf(at+28+4*k,*cast(float*)&src[20+4*k]); // normal
		}
		foreach(i;0..nf){
			auto src=faceData[28*i..28*(i+1)], at=faceOff+faceSize*i;
			foreach(k;0..3) put32(at+4*k,*cast(uint*)&src[4*k]);
			put32(at+12,0x00010001);
			auto t=texNames[*cast(uint*)&src[12]];
			foreach(k;0..4) data[at+16+k]=t[3-k]; // stored reversed
			foreach(k;0..3) putf(at+20+4*k,*cast(float*)&src[16+4*k]);
			put32(at+32,0);
		}
		enforce(parseMRMM(data).faces.length==nf);
		auto vdir=buildPath("extracted","hydro","manalith");
		auto modelPath=buildPath(vdir,"hmnl.MRMM");
		wadManager.files[fixPath(modelPath)]=data;
		bldgModls[hydroManalithTag]=modelPath;
		// textures: found by name when the model is drawn
		import dlib.image.io.png:loadPNG;
		foreach(i,t;texNames){
			auto png=buildPath(dir,i==0?"reefrock.png":"coral.png");
			char[4] key=[t[3],t[2],t[1],t[0]];
			if(exists(png)) textureReplacementsSloppy[key]=loadPNG(png);
		}
		// the building: the Stratos manalith with our model
		auto sb=bldgs[stratos];
		auto h=new BldgHeader; *h=*cast(BldgHeader*)sb.header;
		auto comps=(cast(BldgComponent[])sb.components).dup;
		foreach(ref c;comps) if(c.tag==stratos) c.tag=hydroManalithTag;
		(cast(Bldg[char[4]])bldgs)[hydroManalithTag]=Bldg(h,comps);
		writeln("hydro mod: coral manalith with ",nv," vertices, ",nf," faces");
	}catch(Exception e) stderr.writeln("hydro mod: could not build the manalith: ",e.msg);
}

private double num(JSONValue v){ // JSON numbers may be written as integers or floats
	if(v.type==JSONType.integer) return v.integer;
	if(v.type==JSONType.uinteger) return v.uinteger;
	return v.floating;
}

private char[4] rawTag(string readable){ // tags are stored reversed in the game data
	enforce(readable.length==4,text("tag must have 4 characters: ",readable));
	char[4] r;
	foreach(i;0..4) r[i]=readable[3-i];
	return r;
}

// sea-blue version of an effect texture (used for the effects of abilities Hydro creatures borrow)
import dlib.image:SuperImage;
SuperImage hydroTint(SuperImage img){
	import dlib.image.color:Color4f;
	auto r=img.dup;
	foreach(y;0..r.height) foreach(x;0..r.width){
		auto c=r[x,y];
		auto l=max(c.r,max(c.g,c.b));
		r[x,y]=Color4f(0.2f*l,0.62f*l,l,c.a);
	}
	return r;
}
SuperImage hydroFx(string name){ // a Hydro effect texture from mods/hydro/fx
	import dlib.image.io.png:loadPNG;
	import dlib.image.image:image;
	auto path=buildPath(modsRoot.length?modsRoot:".","mods","hydro","fx",name);
	if(!exists(path)) return image(4,4,4); // without the mod: an empty texture
	return loadPNG(path);
}
// --- new sounds (mods/hydro/sounds): 16-bit mono PCM WAV files become game samples ---
private int numHydroSounds=0;
char[4][string] hydroFxSounds; // ability sound name (mods/hydro/sounds/fx/<name>.wav) -> sample tag
// registers a WAV file as a game sample and returns its tag (or "\0\0\0\0" if it can't be used)
char[4] registerHydroSound(string path){
	import samp:SampHeader;
	try{
		auto data=cast(ubyte[])std.file.read(path);
		enforce(data.length>12&&data[0..4]==cast(ubyte[])"RIFF"&&data[8..12]==cast(ubyte[])"WAVE","not a WAV file");
		uint rate=0; ushort channels=0,bits=0; ubyte[] pcm;
		for(size_t o=12;o+8<=data.length;){
			auto id=cast(char[])data[o..o+4];
			auto size=*cast(uint*)&data[o+4];
			auto body=data[o+8..min(data.length,o+8+size)];
			if(id=="fmt "){ channels=*cast(ushort*)&body[2]; rate=*cast(uint*)&body[4]; bits=*cast(ushort*)&body[14]; }
			else if(id=="data") pcm=body;
			o+=8+size+(size&1);
		}
		enforce(channels==1&&bits==16&&pcm.length,"must be 16-bit mono PCM");
		SampHeader h;
		h.size=cast(uint)pcm.length;
		h.sampleRate=rate;
		h.byteRate=2*rate;
		auto n=numHydroSounds++;
		char[4] readable=['H',"0123456789abcdefghijklmnopqrstuvwxyz"[n/1296%36],"0123456789abcdefghijklmnopqrstuvwxyz"[n/36%36],"0123456789abcdefghijklmnopqrstuvwxyz"[n%36]];
		auto tag=rawTag(readable.idup);
		auto vpath=fixPath(buildPath("extracted","hydro","sounds",readable.idup~".SAMP"));
		wadManager.files[vpath]=(cast(ubyte*)&h)[0..SampHeader.sizeof].dup~pcm;
		samps[tag]=vpath;
		{ import std.process:environment; if(environment.get("HYDRO_DEBUG","").length){ import samp:loadSAMP; auto sm=loadSAMP(vpath); writeln("hydro mod: sound ",baseName(path)," -> ",readable," ",sm.data.length," bytes at ",sm.header.sampleRate," Hz"); } }
		return tag;
	}catch(Exception e){
		stderr.writeln("hydro mod: sound ",path,": ",e.msg);
		return "\0\0\0\0";
	}
}
// a Hydro ability sound, or the given game sound if the mod's file is missing
char[4] hydroSound(string name,char[4] fallback){
	if(auto t=name in hydroFxSounds) return *t;
	return fallback;
}
private void loadHydroFxSounds(string root){
	auto dir=buildPath(root,"sounds","fx");
	if(!exists(dir)||!isDir(dir)) return;
	foreach(string f;dirEntries(dir,"*.wav",SpanMode.shallow)){
		auto tag=registerHydroSound(f);
		if(tag!="\0\0\0\0") hydroFxSounds[baseName(f,".wav")]=tag;
	}
}
// "sounds": {"selected": ["a.wav", ...], "moving": [...], ...} gives a creature its own voice; types
// not listed keep the sounds of the creature it is built on
private void loadCreatureSounds(ref Cre8 c,string dir,JSONValue sounds){
	import sset:Sset,SsetEntry,SoundType;
	static immutable string[] names=["rescued","selected","annoyed","moving","attacking","melee","damaged","death"];
	static immutable SoundType[] types=[SoundType.rescued,SoundType.selected,SoundType.annoyed,SoundType.moving,SoundType.attacking,SoundType.melee,SoundType.damaged,SoundType.death];
	Sset sset;
	if(auto base=c.creatureSSET in ssets) foreach(ref e;base.entries) sset.entries~=SsetEntry(e.type,e.sounds.dup);
	auto root=buildPath(modsRoot.length?modsRoot:".","mods","hydro","sounds");
	foreach(string k,v;sounds.object){
		auto i=names.countUntil(k);
		if(i<0){ stderr.writeln("hydro mod: unknown sound type ",k); continue; }
		char[4][] tags;
		foreach(f;v.array){
			auto path=buildPath(root,f.str);
			if(!exists(path)) path=buildPath(dir,f.str);
			auto tag=registerHydroSound(path);
			if(tag!="\0\0\0\0") tags~=tag;
		}
		if(!tags.length) continue;
		bool found=false;
		foreach(ref e;sset.entries) if(e.type==types[i]){ e.sounds=tags; found=true; }
		if(!found) sset.entries~=SsetEntry(types[i],tags);
	}
	auto n=numHydroSounds++;
	char[4] readable=['S',"0123456789abcdefghijklmnopqrstuvwxyz"[n/1296%36],"0123456789abcdefghijklmnopqrstuvwxyz"[n/36%36],"0123456789abcdefghijklmnopqrstuvwxyz"[n%36]];
	auto tag=rawTag(readable.idup);
	sset.name=tag;
	(cast(Sset[char[4]])ssets)[tag]=sset;
	c.creatureSSET=tag;
}
bool isHydroCreature(char[4] nttTag){ return !!(nttTag in hydroAbilityNames); }

private void setField(ref Cre8 c,string name,JSONValue v){
	static foreach(member;__traits(allMembers,Cre8)){{
		alias T=typeof(__traits(getMember,c,member));
		static if(is(T==ushort)||is(T==uint)||is(T==float)){
			if(name==member){
				__traits(getMember,c,member)=v.type==JSONType.float_?cast(T)v.floating:cast(T)v.integer;
				return;
			}
		}else static if(is(T==char[4])){
			if(name==member){ // a tag such as an ability, readable form; "" clears it
				__traits(getMember,c,member)=v.str.length?rawTag(v.str):"\0\0\0\0";
				return;
			}
		}
	}}
	stderr.writeln("hydro mod: unknown creature stat '",name,"'");
}

// an SXSK animation with more bones: each added bone turns like the bone it copies
// a bone that sways: its rotation in every frame is replaced by a turn about "axis" (SXMD axes)
// of amp*sin(2*pi*cycles*frame/frames+phase), with about one cycle per second of animation
// pow>0 makes the sway a short bump (max(0,sin)^pow) instead of a back-and-forth swing;
// idleOnly limits it to the standing animations (stance and idles)
struct Wave{ int bone; float[3] axis; float amp, phase; float pow=0; bool idleOnly; float period=0; float speed=1; bool moveOnly; }

// copies: source bone of each added bone; shifts: how far ahead in the animation (fraction of
// its length) the added bone takes its rotation from, e.g. 0.5 flaps a second pair of wings in
// counter-time
private ubyte[] addBones(ubyte[] d,int[] copies,Wave[] waves=[],float[] shifts=[],bool idle=false,bool slowBob=false,bool moving=false){
	import std.math:sin,cos,PI,round,pow,ceil;
	auto numFrames=*cast(ushort*)&d[2];
	auto nb=*cast(ushort*)&d[8];
	auto nn=nb+copies.length;
	// a standing animation with a slow wave (period, in seconds at 30 frames per second) is
	// played several times over, so the wave has room for one long, smooth cycle
	int reps=1;
	if(idle) foreach(w;waves) if(w.idleOnly&&w.period>0)
		reps=max(reps,cast(int)round(w.period*30/numFrames));
	if(slowBob) reps=max(reps,2); // the body's up-and-down motion is played at half speed over two cycles
	auto total=numFrames*reps;
	enforce(total<=ushort.max);
	auto headerEnd=12+12*total;
	auto r=new ubyte[](headerEnd);
	r[0..12]=d[0..12];
	*cast(ushort*)&r[2]=cast(ushort)total;
	*cast(ushort*)&r[8]=cast(ushort)nn;
	foreach(f;0..total){
		auto f0=f%numFrames;
		r[12+12*f..12+12*f+12]=d[12+12*f0..12+12*f0+12];
		if(slowBob){
			double x=cast(double)f*numFrames/total;
			int i0=cast(int)x%numFrames, i1=(i0+1)%numFrames;
			double t=x-cast(int)x;
			foreach(ci;0..3){
				auto a=*cast(short*)&d[12+12*i0+2+2*ci], b=*cast(short*)&d[12+12*i1+2+2*ci];
				*cast(short*)&r[12+12*f+2+2*ci]=cast(short)round(a+(b-a)*t);
			}
		}
		auto off=*cast(uint*)&d[12+12*f0+8];
		// the frame's rotations, then the added bones' (copied from the original data), then
		// the waves on top, on any bone including the added ones
		auto rot=(cast(short[4][])d[off..off+8*nb]).dup;
		foreach(k,c;copies){
			auto sh=k<shifts.length?shifts[k]:0.0f;
			auto g=(f0+cast(int)round(sh*numFrames))%numFrames;
			auto og=*cast(uint*)&d[12+12*g+8];
			rot~=(cast(short[4][])d[og+8*c..og+8*c+8])[0];
		}
		foreach(w;waves){
			if(w.bone>=rot.length) continue;
			if(w.idleOnly&&!idle) continue;
			if(w.moveOnly&&!moving) continue; // run and walk only
			double cycles=w.idleOnly&&w.period>0?1.0:max(1.0,round(total/30.0*w.speed));
			auto sv=sin(2*PI*cycles*f/total+w.phase);
			auto a=w.pow>0?w.amp*pow(max(0.0,sv),w.pow):w.amp*sv;
			foreach(ci;0..3) rot[w.bone][ci]=cast(short)round(16000*w.axis[ci]*sin(a/2));
			rot[w.bone][3]=cast(short)round(16000*cos(a/2));
		}
		*cast(uint*)&r[12+12*f+8]=cast(uint)r.length;
		r~=cast(ubyte[])rot;
	}
	return r;
}

// rig.json "poses": holds some bones in the pose of one frame of another animation through the middle
// of an animation, easing in and out (e.g. the Tidecaller raises its staff while it shoots)
private struct Pose{ int[] slots; int from; float frame; int[] bones; float easeIn=0.25f, easeOut=0.8f; }
private void blendPose(ubyte[] r,ubyte[] src,Pose p){
	import std.math:sqrt,round;
	auto n=*cast(ushort*)&r[2], nb=*cast(ushort*)&r[8];
	auto sn=*cast(ushort*)&src[2], snb=*cast(ushort*)&src[8];
	auto sf=cast(int)round(p.frame*(sn-1));
	auto soff=*cast(uint*)&src[12+12*sf+8];
	auto target=cast(short[4][])src[soff..soff+8*snb];
	foreach(f;0..n){
		float t=n>1?float(f)/(n-1):0.0f;
		float w=1.0f;
		static float smooth(float x){ x=max(0.0f,min(1.0f,x)); return x*x*(3-2*x); }
		if(t<p.easeIn) w=smooth(t/p.easeIn);
		else if(t>p.easeOut) w=smooth((1-t)/(1-p.easeOut));
		auto off=*cast(uint*)&r[12+12*f+8];
		auto rot=cast(short[4][])r[off..off+8*nb];
		foreach(b;p.bones){
			if(b>=nb||b>=snb) continue;
			float[4] a,q;
			foreach(c;0..4){ a[c]=rot[b][c]; q[c]=target[b][c]; }
			float dot=0,la=0;
			foreach(c;0..4){ dot+=a[c]*q[c]; la+=a[c]*a[c]; }
			if(dot<0) foreach(c;0..4) q[c]=-q[c];
			float[4] m; float lm=0;
			foreach(c;0..4){ m[c]=a[c]*(1-w)+q[c]*w; lm+=m[c]*m[c]; }
			auto sc=sqrt(la)/max(1e-6f,sqrt(lm));
			foreach(c;0..4) rot[b][c]=cast(short)round(m[c]*sc);
		}
	}
}

private struct PendingAnim{ char[4] key; string path; }
private PendingAnim[] pendingAnims;

void loadHydroMod(){
	if(!wadManager) return;
	auto base=buildPath(modsRoot.length?modsRoot:".","mods","hydro","creatures");
	if(!exists(base)||!isDir(base)) return;
	addMaelstrom();
	addHydroSpells(buildPath(modsRoot.length?modsRoot:".","mods","hydro"));
	addHydroManalith(buildPath(modsRoot.length?modsRoot:".","mods","hydro","manalith"));
	{ import std.process:environment; if(environment.get("HYDRO_DEBUG","").length){ // what the manaliths are made of
		import mrmm:parseMRMM;
		if(auto st=SpellTag.manalith in strcs) foreach(i,bt;st.structure){
			if(auto b=bt in bldgs){
				writeln("hydro mod: manalith ",i," building ",bt," components ",b.components.length);
				foreach(ref c;b.components){
					writeln("  component ",c.kind," ",c.tag," at ",c.x,",",c.y,",",c.z," model ",bldgModls.get(c.tag,"?"));
					if(auto path=c.tag in bldgModls){ auto m=parseMRMM(readFile(*path)); float[3] lo=float.max,hi=-float.max; foreach(ref v;m.vertices[0]) foreach(k;0..3){ lo[k]=min(lo[k],v.pos[k]); hi[k]=max(hi[k],v.pos[k]); } writeln("   frames ",m.vertices.length," verts ",m.vertices[0].length," faces ",m.faces.length," bounds ",lo," ",hi," hitboxes ",m.hitboxes.length," tex ",m.faces.length?m.faces[0].textureName:""); }
				}
			}
		}
	}}
	loadHydroFxSounds(buildPath(modsRoot.length?modsRoot:".","mods","hydro"));
	foreach(string dir;dirEntries(base,SpanMode.shallow)){
		auto def=buildPath(dir,"creature.json");
		if(!exists(def)) continue;
		try loadCreature(dir,parseJSON(readText(def)));
		catch(Exception e) stderr.writeln("hydro mod: could not load ",dir,": ",e.msg);
	}
	// register the copied animations only now, and in front: the originals keep winning any ties
	foreach(p;pendingAnims) saxsAnims[p.key]=p.path~saxsAnims.get(p.key,[]);
	pendingAnims=[];
	{ import std.process:environment; if(environment.get("HYDRO_STATS","").length) dumpCreatureStats(); }
	{ import std.process:environment; if(environment.get("HYDRO_SPELLS","").length) dumpSpellbookSlots(); }
}

// HYDRO_SPELLS=1: prints every god's creatures and spells in spellbook order (for the launcher's spellbook editor)
private void dumpSpellbookSlots(){
	import ntts:God;
	import std.traits:EnumMembers;
	foreach(god;EnumMembers!God){
		foreach(kind,list;[creatureSpells[god],normalSpells[god]]) foreach(i,tag;list){
			char[4] t=tag; auto c=t in cre8s; auto sp=t in spels;
			auto name=c?texts.get(c.name,""):sp?texts.get(sp.name,""):"";
			writefln("SPELL %s %s %d %s %s",god,kind==0?"creature":"spell",i,[t[3],t[2],t[1],t[0]],name);
		}
	}
	foreach(i,tag;structureSpells){ char[4] t=tag; auto st=t in strcs; writefln("SPELL none structure %d %s %s",i,[t[3],t[2],t[1],t[0]],st?texts.get(st.name,""):""); }
}

// HYDRO_STATS=1: prints the combat numbers of the creatures of Pyro, James, Stratos and Hydro
private void dumpCreatureStats(){
	void spel(string what,char[4] t){
		if(t=="\0\0\0\0") return;
		if(auto q=t in spels) writefln("    %s %s \"%s\" amount %s amount2 %s cooldown %.2fs range %s effectRange %s damageRange %s duration %s speed %s",what,t,texts.get(q.name,""),q.amount,q.amount2,q.cooldown/60.0f,q.range,q.effectRange,q.damageRange,q.duration,q.speed);
		else writefln("    %s %s (not a spell)",what,t);
	}
	static foreach(g;["persephone","pyro","james","stratos","charnel","hydro"]){{
		writeln("== ",g);
		foreach(tag;mixin(g~"Creatures")){
			auto c=tag in cre8s; if(!c) continue;
			writefln("  %s \"%s\" mana %s souls %s health %s meleeStrength %s meleeAccuracy %s meleeEvasion %s rangedAccuracy %s meleeRes %s directSpellRes %s splashSpellRes %s directRangedRes %s splashRangedRes %s flags %x",
				tag,texts.get(c.name,""),c.manaCost,c.souls,c.health,c.meleeStrength,c.meleeAccuracy,c.meleeEvasion,c.rangedAccuracy,c.meleeResistance,c.directSpellResistance,c.splashSpellResistance,c.directRangedResistance,c.splashRangedResistance,c.behaviorFlags);
			spel("ability0",c.ability0); spel("ability1",c.ability1); spel("ability2",c.ability2);
			spel("passive",c.passiveAbility); spel("passive2",c.passiveAbility2);
			if(auto r=tag in hydroRangedStats) writefln("    own ranged attack numbers: amount %s range %s",r.amount,r.range);
		}
	}}
}

private void loadCreature(string dir,JSONValue js){
	{ // rig.json (written by the model generator): skeleton changes and borrowed parts
		auto rig=buildPath(dir,"rig.json");
		if(exists(rig)) foreach(string k,v;parseJSON(readText(rig)).object) js[k]=v;
	}
	auto tag=rawTag(js["tag"].str);
	auto model=js["model"].str;
	auto rawModel=rawTag(model);
	auto baseTag=rawTag(js["base"].str);
	enforce(baseTag in cre8s,text("unknown base creature ",js["base"].str));
	Cre8 c=cre8s[baseTag];
	// virtual files: placed next to the base creature's model so that its animations are found
	auto baseModel=saxsModls[c.saxsModel];
	auto vdir=buildPath(dirName(baseModel),model~".FLDR"); // inside the base model's folder, so its animations are the closest match
	foreach(string f;dirEntries(dir,SpanMode.shallow)){
		auto name=baseName(f);
		if(name.endsWith(".SXMD")||name.endsWith(".SAXS")||name.endsWith(".SXTX"))
			wadManager.files[fixPath(buildPath(vdir,name))]=cast(ubyte[])read(f);
	}
	auto sxmd=buildPath(vdir,model~".SXMD");
	enforce(fileExists(sxmd),text("missing ",model,".SXMD"));
	uint n=*cast(uint*)&(readFile(baseModel)[28]); // bones including the root
	{ // the mod does not ship the skeleton: copy it, the reference points and hitbox bones from the base creature
		auto mine=wadManager.files[fixPath(sxmd)], base=readFile(baseModel);
		uint mb=*cast(uint*)&mine[32], bb=*cast(uint*)&base[32];
		// rig.json may move or resize bones of the base skeleton ("boneScale", "boneShift", in
		// SXMD units) and add bones after it ("extraBones"), each animated like the bone it copies.
		// Everything is computed here from the player's own game files.
		uint m=*cast(uint*)&mine[28];
		JSONValue[] extras;
		if(auto eb="extraBones" in js) extras=eb.array;
		enforce(m==n+extras.length&&m<=32,"skeleton does not match the base creature");
		mine[mb..mb+n*140]=base[bb..bb+n*140];
		float* pos(size_t i){ return cast(float*)&mine[mb+i*140]; }
		ushort* parent(size_t i){ return cast(ushort*)&mine[mb+i*140+12]; }
		float[3] world(size_t i){
			float[3] w=0;
			for(;i>0;i=*parent(i)) foreach(ci;0..3) w[ci]+=pos(i)[ci];
			return w;
		}
		float[3] vec(JSONValue v){ float[3] r; foreach(ci;0..3) r[ci]=cast(float)num(v.array[ci]); return r; }
		if(auto bs="boneScale" in js) foreach(string k,v;bs.object){
			auto i=to!int(k); enforce(1<=i&&i<n);
			foreach(ci;0..3) pos(i)[ci]*=cast(float)num(v);
		}
		if(auto bs="boneShift" in js) foreach(string k,v;bs.object){
			auto i=to!int(k); enforce(1<=i&&i<n);
			auto d=vec(v);
			foreach(ci;0..3) pos(i)[ci]+=d[ci];
		}
		if(auto rb="rebone" in js) foreach(e;rb.array){ // base bones re-used elsewhere (e.g. legs as a tail)
			auto i=cast(int)e["bone"].integer, par=cast(int)e["parent"].integer;
			enforce(1<=i&&i<n&&0<=par&&par<i,"bad rebone");
			*parent(i)=cast(ushort)par;
			auto d=vec(e["pos"]);
			foreach(ci;0..3) pos(i)[ci]=d[ci];
		}
		foreach(k,e;extras){
			auto i=n+k, src=cast(int)e["copy"].integer, par=cast(int)e["parent"].integer;
			enforce(1<=src&&src<n&&0<=par&&par<i,"bad extra bone");
			mine[mb+i*140..mb+(i+1)*140]=0;
			*parent(i)=cast(ushort)par;
			if(par>=n){ // hangs off another added bone: the source's own offset, scaled
				auto sc="scale" in e?cast(float)num(e["scale"]):1.0f;
				foreach(ci;0..3) pos(i)[ci]=pos(src)[ci]*sc;
			}else{ // hangs off a base bone: where the source is, moved by "shift"
				auto ws=world(src), wp=world(par);
				float[3] d=0;
				if("shift" in e) d=vec(e["shift"]);
				foreach(ci;0..3) pos(i)[ci]=ws[ci]-wp[ci]+d[ci];
			}
		}
		auto msaxs=wadManager.files[fixPath(buildPath(vdir,model~".SAXS"))], bsaxs=readFile(baseModel[0..$-5]~".SAXS");
		msaxs[0x18..0x58]=bsaxs[0x18..0x58];
		msaxs[0xB8..0xB8+40]=bsaxs[0xB8..0xB8+40];
		foreach(i;n..40) msaxs[0xB8+i]=0; // added bones are not hitbox bones
	}
	saxsModls[rawModel]=sxmd;
	if("rawUV" in js && js["rawUV"].boolean) rawUVModels[fixPath(sxmd)]=true;
	if(auto nd="noDisplacement" in js){
		int[] l;
		foreach(v;nd.array) l~=cast(int)v.integer;
		noDispModels[fixPath(sxmd)]=l;
	}
	if("borrowParts" in js){ // a part number, or {"part":p, "bones":{"base bone":our bone}, "scale":s}
		BorrowedPart[] parts;
		foreach(v;js["borrowParts"].array){
			if(v.type==JSONType.integer){ parts~=BorrowedPart(cast(int)v.integer); continue; }
			BorrowedPart b;
			b.part=cast(int)v["part"].integer;
			if(auto bones="bones" in v) foreach(string k,t;bones.object) b.boneMap[to!int(k)]=cast(int)t.integer;
			if(auto sc="scale" in v) b.scale=cast(float)num(*sc);
			if(auto rb="rebind" in v) foreach(string k,t;rb.object) b.rebind[to!int(k)]=cast(int)t.integer;
			parts~=b;
		}
		borrowedParts[sxmd]=BorrowedParts(baseModel,parts);
	}
	Wave[] waves;
	if(auto wv="wave" in js) foreach(v;wv.array){
		Wave w;
		w.bone=cast(int)v["bone"].integer;
		foreach(ci;0..3) w.axis[ci]=cast(float)num(v["axis"].array[ci]);
		w.amp=cast(float)num(v["amp"]);
		w.phase=cast(float)num(v["phase"]);
		if("pow" in v) w.pow=cast(float)num(v["pow"]);
		if("idleOnly" in v) w.idleOnly=v["idleOnly"].boolean;
		if("period" in v) w.period=cast(float)num(v["period"]);
		if("speed" in v) w.speed=cast(float)num(v["speed"]);
		if("moveOnly" in v) w.moveOnly=v["moveOnly"].boolean;
		waves~=w;
	}
	int[] boneCopies;
	float[] boneShifts;
	if(auto eb="extraBones" in js) foreach(v;eb.array){
		boneCopies~=cast(int)v["copy"].integer;
		boneShifts~="frameShift" in v?cast(float)num(v["frameShift"]):0.0f;
	}
	Pose[] poses;
	if(auto ps="poses" in js) foreach(v;ps.array){
		Pose p;
		foreach(x;v["slots"].array) p.slots~=cast(int)x.integer;
		p.from=cast(int)v["from"].integer;
		p.frame=cast(float)num(v["frame"]);
		foreach(x;v["bones"].array) p.bones~=cast(int)x.integer;
		if("in" in v) p.easeIn=cast(float)num(v["in"]);
		if("out" in v) p.easeOut=cast(float)num(v["out"]);
		poses~=p;
	}
	{ // give the new model its own copies of the base creature's animations, so they are always the ones found
		import std.ascii:toLower;
		string animSource(char[4] animID){ // the closest animation file that fits the skeleton (same bone count)
			char[4] key=animID;
			foreach(ref ch;key) ch=toLower(ch);
			string src=null;
			size_t best=0;
			foreach(cand;saxsAnims.get(key,[])){
				if(!fileExists(cand)) continue;
				auto d=readFile(cand);
				if(d.length<10||*cast(ushort*)&d[8]!=n) continue;
				size_t cur=0;
				while(cur<cand.length&&cur<baseModel.length&&cand[cur]==baseModel[cur]) cur++;
				if(cur>=best){ best=cur; src=cand; }
			}
			return src;
		}
		foreach(slot,animID;c.animations.animations){
			bool idle=slot<=4; // stance, idle0..3
			bool slowBob=false;
			if(auto sb="slowBob" in js) foreach(v;sb.array) if(v.integer==slot) slowBob=true;
			if(animID=="\0\0\0\0"||animID=="rezW"||animID[0..2]=="00") continue;
			char[4] key=animID;
			foreach(ref ch;key) ch=toLower(ch);
			// the closest animation file that fits the skeleton (same bone count)
			string src=null;
			size_t best=0;
			foreach(cand;saxsAnims.get(key,[])){
				if(!fileExists(cand)) continue;
				auto d=readFile(cand);
				if(d.length<10||*cast(ushort*)&d[8]!=n) continue;
				size_t cur=0;
				while(cur<cand.length&&cur<baseModel.length&&cand[cur]==baseModel[cur]) cur++;
				if(cur>=best){ best=cur; src=cand; }
			}
			if(!src.length){
				import std.process:environment;
				if(environment.get("HYDRO_DEBUG","").length) foreach(cand;saxsAnims.get(key,[])) stderr.writeln("hydro mod: ",model," ",key," candidate ",cand," bones ",fileExists(cand)?*cast(ushort*)&readFile(cand)[8]:-1," need ",n);
				continue;
			}
			auto dst=fixPath(buildPath(vdir,text(cast(uint)(key[0]|key[1]<<8|key[2]<<16|key[3]<<24),".SXSK")));
			if(dst in wadManager.files) continue;
			wadManager.files[dst]=boneCopies.length||waves.length||slowBob?addBones(readFile(src),boneCopies,waves,boneShifts,idle,slowBob,slot==6||slot==24):readFile(src).dup;
			foreach(p;poses) if(p.slots.canFind(cast(int)slot)){
				auto from=animSource(c.animations.animations[p.from]);
				if(from.length) blendPose(wadManager.files[dst],readFile(from),p);
			}
			auto skel=src[0..$-5]~".SKEL";
			if(fileExists(skel)) wadManager.files[dst[0..$-5]~".SKEL"]=readFile(skel);
			pendingAnims~=PendingAnim(key,dst);
			hydroAnimOwner[dst]=fixPath(sxmd);
		}
	}
	c.name=tag;
	c.saxsModel=rawModel;
	if("stats" in js) foreach(string k,v;js["stats"].object) setField(c,k,v);
	if("abilityNames" in js) foreach(string k,v;js["abilityNames"].object) hydroAbilityNames[tag][rawTag(k)]=v.str;
	if("boltProjectile" in js&&js["boltProjectile"].boolean) hydroBoltCreatures[tag]=true;
	if("startsFlying" in js&&js["startsFlying"].boolean) hydroStartsFlying[tag]=true;
	if("shootFrom" in js) hydroShootBone[tag]=cast(int)js["shootFrom"]["bone"].integer;
	if("rangedAttack" in js){ // this creature's own numbers for the ranged attack it shares with others
		HydroRangedStats r;
		auto ra=js["rangedAttack"];
		if("amount" in ra) r.amount=ra["amount"].get!float;
		if("range" in ra) r.range=ra["range"].get!float;
		hydroRangedStats[tag]=r;
	}
	if("sounds" in js) loadCreatureSounds(c,dir,js["sounds"]);
	// "animationSlots": {"22": 11} plays the animation of slot 11 in slot 22 (e.g. a different
	// animation for an ability); slots: 0 stance, 11-13 attacks, 22-23 shooting/abilities
	{ import animations:Animations; static assert(Animations.disoriented.offsetof/4==56&&Animations.rise.offsetof/4==46); } // 56: summoning pose
	if("animationSlots" in js) foreach(string k,v;js["animationSlots"].object)
		c.animations.animations[to!int(k)]=c.animations.animations[cast(int)num(v)];
	(cast(string[char[4]])texts)[tag]=js["name"].str;
	auto icon=buildPath(dir,"icon.png");
	if(exists(icon)){
		import dlib.image.io.png:loadPNG;
		textureReplacementsSloppy[rawModel]=loadPNG(icon); // the icon's tag comes from its file name
		icons[tag]=buildPath(vdir,model~".ICON");
		c.icon=tag;
	}
	(cast(Cre8[char[4]])cre8s)[tag]=c;
	writeln("hydro mod: loaded creature ",js["name"].str);
	{ import std.process:environment; if(environment.get("HYDRO_DEBUG","").length) writeln("hydro mod: model ",sxmd," stance ",getSaxsAnim(baseModel,c.animations.stance1)); }
}
