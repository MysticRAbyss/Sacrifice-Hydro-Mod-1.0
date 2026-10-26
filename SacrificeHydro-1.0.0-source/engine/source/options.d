// copyright © tg
// distributed under the terms of the gplv3 license
// https://www.gnu.org/licenses/gpl-3.0.txt

import ntts: God;
import hotkeys_: Hotkeys;

enum GameMode{
	scenario,
	skirmish,
	slaughter,
	domination,
	soulHarvest,
}

enum God randomGod=cast(God)-1; // god option value: pick a random god

struct Options{
	// graphics options
	int width=1280, height=720;
	bool detectResolution=false;
	bool resizableWindow=false;
	bool enableFullscreen=false;
	bool captureMouse=false;
	bool focusOnStart=false;
	float scale=1.0f;
	bool scaleToFit=true;
	bool centerViewport=true;
	float aspectDistortion=1.2f;
	float sunFactor=1.0f;
	float ambientFactor=1.0f;
	int shadowMapResolution=1024;
	bool enableWidgets=true;
	bool enableMapBottom=true;
	bool enableFog=false;
	bool enableSSAO=false;
	bool enableGlow=true;
	float glowBrightness=0.5;
	bool enableAntialiasing=true;
	// Hydro mod: picture and frame rate
	bool vsync=true;          // with adaptive vsync where the driver has it (late frames are not held back to 30)
	int fpsLimit=0;           // 0: no limit (besides vsync)
	float anisotropy=16.0f;   // anisotropic texture filtering (1: off)
	bool originalColors=true; // as the original game: no tone curve; --no-original-colors: the engine's filmic (ACES) look
	float exposure=0.7f;      // brightness
	float saturation=1.5f;    // colour strength
	float contrast=1.2f;
	int cursorSize=-1;
	bool freetypeFonts=true;
	bool printFps=false;
	string hdTextures=null; // Hydro mod: folder of Sacrifice HD's high-resolution wads ("--hd": its usual place), used in place of the originals
	// audio options
	float volume=1.0f;
	float musicVolume=1.0f;
	float soundVolume=1.0f;
	bool advisorHelpSpeech=true;
	// input options
	string hotkeyFilename="";
	Hotkeys hotkeys;
	int doubleClickSpeed=20;
	bool xmenu=true;
	int xmenuTime=15;
	float xmenuHiddenScale=0.4f;
	// int keyRepeatStarDelay=50; // TODO?
	// int keyRepeatDelay=10;
	// float mouseSpeed=1.0f; // TODO?
	float cameraMouseSensitivity=1.0f;
	float mouseWheelSensitivity=1.0f;
	bool windowScrollX=true;
	bool windowScrollY=true;
	float windowScrollXFactor=1.0f;
	float windowScrollYFactor=1.0f;
	// bool mouseWheelZoom=true; // TODO
	bool debugHotkeys=false;
	// player-specific settings
	God god;
	Settings settings;
	bool randomGods=false;
	bool randomSpellbook=false;
	// global settings
	bool noMap=false;
	int[] teamSizes;
	bool mirrorMatch=false;
	bool shuffleSides=false;
	bool shuffleTeams=false;
	bool shuffleAltars=false;
	bool shuffleSlots=false;
	bool randomWizards=false;
	bool randomSpellbooks=false;
	// host game setup
	int[] sides;
	int[] botSlots;
	God[] gods;
	God[] botGods;
	string[] botNames;
	float damageFactor=1.0f;
	float[] damageFactors;
	float botDamageFactor=1.0f;
	float[] botDamageFactors;
	// global game mode settings:
	GameMode gameMode=GameMode.skirmish;
	int gameModeParam;
	int replicateCreatures=1;
	int protectManafounts=0;
	bool terrainSineWave=false;
	bool fogOfWar=false;
	bool fogOfWar3d=false;
	bool alliedVision=true;
	bool alliedBeamVision=true;
	bool collectAlliedSouls=false;
	bool friendlyFire=true;
	bool alwaysGib=false;
	float gibChance=1.0f; // Hydro mod: chance factor for a creature to be gibbed when it is killed (0: never)
	float friendlyDamage=1.0f; // Hydro mod: damage factor for hits on your own and your allies' units
	bool allowObservers=true; // Hydro mod: the host lets observers join
	bool customSpellbooks=true; // Hydro mod: the host lets players bring their own spellbooks
	bool spellbookGiven=false; // Hydro mod: --spellbook was used
	immutable(SpellSpec)[] pendingSpellbook; // Hydro mod: the --spellbook (names resolved once the game data is loaded)
	bool mapWizards=false;
	bool mapCreatures=true;
	bool mapSouls=true;
	bool revealBlueSouls=true;
	bool dimUnexplored=false;
	bool randomCreatureScale=false;
	bool enableDropSoul=true;
	bool targetDroppedSouls=false;
	bool charmNeutralCreatures=false;
	bool enableParticles=true;
	bool greenAllySouls=false;
	bool fasterStandupTimes=true;
	bool fasterCastingTimes=true;
	bool alwaysOnMinimap=false;
	bool betaPatchBots=false;
	bool aiSides=true;
	int delayStart=0;
	// zerotier
	string zerotierIdentity="zerotier-identity";
	ulong zerotierNetwork=0;
	// multiplayer
	bool host=false;
	int numSlots=1;
	string joinIP="";
	bool testLag=false;
	bool testRenderDelay=false;
	bool advertiseGame=true;
	bool dumpTraffic=false;
	bool dumpNetworkStatus=false;
	bool dumpNetworkSettings=false;
	bool checkDesynch=true;
	bool stutterOnDesynch=true; // TODO: disable this by default once that works
	bool kickOnDesynch=false;
	bool kickObserversOnDesynch=true;
	bool logDesynch=true;
	bool nudgeTimers=true;
	bool dropOnTimeout=true;
	bool resetOnRejoin=true;
	bool synchronizeObserverChat=true;
	bool synchronizeLevel=true;
	bool synchronizeSouls=true;
	bool synchronizeLevelBounds=true;
	bool synchronizeXpRate=true;
	bool useMapSettings=false;
	// recording and playback
	string recordingFilename="";
	string recordingFolder="";
	bool compressRecording=true;
	int logCore=0;
	string playbackFilename="";
	string continueFilename="";
	int continueFrame=int.min;
	bool quickContinue=true;
	// asset export
	int exportFrame=0;
	string exportFolder="sacengine-exports";
	alias settings this;
	// compatibility options / testing
	bool lowmem=false;
	bool enableReadFromWads=true;
}

struct SpellSpec{
	int level;
	char[4] tag;
}
struct Settings{
	string commit;
	string map="";
	int mapHash=0;
	string name="";
	bool observer=false;
	bool observerChat=true;
	bool observerMinimap=false;
	int slot=-1;
	int team=-1;
	char[4] wizard="";
	immutable(SpellSpec)[] spellbook;
	int level=9;
	int souls=12;

	int minLevel=1;
	int maxLevel=9;
	float xpRate=1.0f;

	bool pauseOnDrop=true;
	bool refuseGreenSouls=false;
	// (remember to update Network!B.initSlots)
	// TODO: automate
}
