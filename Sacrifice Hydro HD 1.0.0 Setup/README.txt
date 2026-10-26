Sacrifice Hydro HD 1.0.0
========================

Hydro, god of water, joins the five gods of Sacrifice as a sixth god - together with the
high-resolution textures of Sacrifice HD, built from your own copy of the game.

THIS MOD NEEDS THE ORIGINAL GAME. None of Sacrifice's files are included: the mod uses the files
of your copy of Sacrifice (GOG or Steam version) and does not work without them. Your game files
are never changed.

Requirements
- Sacrifice (GOG or Steam version, or an installed CD version with its "data" and "maps" folders).
- Windows 10 or 11, 64-bit, and a graphics card with OpenGL 4.x.
- Microsoft Visual C++ Redistributable 2015-2022 (x64). Most PCs have it; if not, the setup offers
  the download: https://aka.ms/vs/17/release/vc_redist.x64.exe
- About 300 MB of free space for the HD texture cache. Building it uses your graphics card
  through Vulkan (Real-ESRGAN); without Vulkan it uses a slower built-in upscaler.

Install
- Unzip the Setup download (right-click, "Extract All...") and run
  "Install Sacrifice Hydro HD.exe" in the unzipped folder.
  It finds your Sacrifice folder (or lets you pick it), checks that the original game is there
  and installs the mod into Sacrifice\SacrificeHydroHD. It also makes Desktop and Start menu
  shortcuts and an entry in Windows' Apps & features to uninstall it.
- Updating: run the newer setup over the old one. Your settings, key bindings and saved games
  are kept.
- Manual install (instead of the setup): unzip the "SacrificeHydroHD" folder into your Sacrifice
  folder, so that you get ...\Sacrifice\SacrificeHydroHD\Sacrifice Hydro Launcher.exe next to
  Sacrifice's "data" and "maps" folders. To remove it, delete that folder.
- Uninstall: Apps & features, the Start menu, or
  "Uninstall Sacrifice Hydro HD.exe" in the mod's folder. It asks whether to keep
  your settings and saved games.
- The HD textures are built the first time you press Play (a few minutes, once). They are kept in
  %LOCALAPPDATA%\SacrificeHD\cache, which Sacrifice HD uses too; delete that folder to free the
  space if you remove both.

Play
- Start "Sacrifice Hydro HD" from the Desktop or Start menu (Sacrifice Hydro Launcher.exe).
  The launcher is the game's start menu:
    Single player   you against 1 to 5 computer wizards, each with the god you choose (or Random),
                    on any skirmish map, including maps you add to Sacrifice's maps folder.
    Host a game     a multiplayer game; the others join with your IP address.
    Join a game     enter the host's address. Everyone needs the same version of the mod.
  Choose your name, god (Hydro, the five gods of Sacrifice, or Random) and wizard, then Play.
- Game settings... - the original Game Settings: game type (Skirmish, Slaughter, Soul Harvest,
  Domination), the rules (Random Altar, Observers, Custom Spellbooks, Allied Vision, Friendly Fire,
  Always Gib, No Map Creatures/Souls, ...) and the settings (starting souls, levels, AI damage,
  gib chance, XP rate, friendly damage).
- Spellbook... - your own spellbook (the original game unlocked it after the campaign).
- Options... - screen, sound and mouse settings; the Picture page (colours - vivid like the
  original game or filmic -, brightness, colour strength, contrast, texture sharpness, VSync and
  a frame rate limit); and the Key bindings page.
- Load game... - continue a saved game.
- In a game: Esc opens the game menu (Save Game, Load Game, Surrender, Resume Game); F1 is a quick
  save and F2 a quick load. Saved games are in the mod's "saves" folder (single player only).
  When every enemy is defeated the Victory window appears; OK closes the game.

Hydro
Creatures (level, souls, health, attack, ability):
  Tidekin        1  1   800  melee 160            Shell Up - its shell halves damage taken
  Reefstalker    1  1   545  Water Bow            a naga archer
  Mistwing       1  1   450  Brine Spray          Bubble Shield - shields an ally; flies
  Mantis Shrimp  2  2  1235  Slowing Bolt         long range; slows what it hits
  Brineclaw      3  2  1800  melee 280            Crushing Grip - drags an enemy into its claws
  Tidecaller     4  2  1300  Water Bolt 250       Tidal Mend - heals friendly creatures nearby
  Puffy          5  3  1900  Water Bolt 400       Toxic Spines - poisons every enemy around it
  Reefbreaker    6  3  1450  Whirlpool Mortar 500 a hermit-crab cannon
  Abyss Guardian 7  4  4500  melee 700            Tidal Slam - knocks enemies down around it
  Leviathan      9  5  5200  melee 800            Tidal Breath (water blast), Maelstrom
Spells:
  1  Water Jet       a jet of water: Fireball damage and a knock-back
  2  Bubble Armor    a bubble that cuts damage and heals the wizard
  3  Undertow        a whirlpool that snares a creature and slows those around it
  4  Healing Spring  a spring that heals your side standing in it
  5  Geyser          a column of water that damages and throws enemies
  6  Tidal Wall      a waterfall wall that stops missiles and slows enemies
  7  Acid Rain       corrosive rain that eats through protections
  8  Tsunami         a huge wave that sweeps enemies away
  8  Kraken          a giant whirlpool with tentacles that seize anything - friend or foe
- Hydro has its own manalith (coral-crowned reef rock), sounds, voices and spell effects.
- The new creatures are built on the skeletons of the original ones and move with their
  animations.

Also new in the engine
- The original Game Settings, custom spellbooks, the Victory and Defeat windows, the in-game
  menu with saving, loading and surrender, and a key binding editor.
- The HD textures of Sacrifice HD (upscaled with Real-ESRGAN from your copy of the game).
- If you also use the separate Sacrifice HD mod, both share the texture cache; keep both at
  the versions released together so the cache is not rebuilt.

Known limits
- Mission Objectives, Speech History and the in-game Options are not available in skirmish
  games; the campaign is not part of the mod. View Summary after a defeat is not available yet.
- Saving and loading work in single player games.
- Multiplayer: every player needs the same version of the mod.

Files
- Sacrifice Hydro Launcher.exe - start here. Hydro.exe - the game (sacengine with the Hydro mod).
- mods\hydro - Hydro's creatures, spells, manalith, effects and sounds.
- settings.txt - the options (the launcher edits it). hotkeys.txt - the key bindings.
- source\SacrificeHydro-source.zip - the complete source code (GPL-3.0).
- HDTextures - the Sacrifice HD texture builder.

Credits and licences
- Sacrifice by Shiny Entertainment (Interplay). This is an unofficial fan mod; it includes none of
  the game's files.
- sacengine by tg and contributors (https://github.com/tg-2/sacengine), GPL-3.0. Hydro.exe is a
  modified version and is therefore also GPL-3.0; its full source is in the source folder.
- dagon and dlib by Timur Gafarov et al. (Boost), SDL2 (zlib), FreeType (FreeType licence),
  OpenAL Soft (LGPL, unmodified DLL).
- The launcher and setup use walk and win by The Walk Authors (BSD-3-Clause), go-ole (MIT),
  golang.org/x/sys
  (BSD-3-Clause) and govaluate (MIT).
- Creature sounds made from CC0 packs: "80 CC0 creature SFX" and "80 CC0 creature SFX #2" by
  rubberduck and "Monster SFX pack" by Ogrebane (opengameart.org) - with thanks.
- Spoken lines synthesised with Piper by Michael Hansen / Rhasspy (MIT) using the
  "en-us-libritts-high" voice, trained on LibriTTS by Heiga Zen et al. (CC BY 4.0,
  http://www.openslr.org/60/), from LibriVox public-domain audiobooks.
- Texture upscaling: Real-ESRGAN / realesrgan-ncnn-vulkan by Xintao Wang et al.
  (BSD-3-Clause / MIT).
- Licence texts are in the licenses folder and HDTextures\licenses.

