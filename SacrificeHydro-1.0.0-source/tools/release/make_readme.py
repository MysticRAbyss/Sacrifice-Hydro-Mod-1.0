# Writes the release README.txt of Sacrifice: Hydro (edition "plain" or "hd")
import sys
edition, version, out = sys.argv[1], sys.argv[2], sys.argv[3]
hd = edition == "hd"
title = "Sacrifice Hydro HD" if hd else "Sacrifice: Hydro"
folder = "SacrificeHydroHD" if hd else "SacrificeHydro"
setup = f"Install {'Sacrifice Hydro HD' if hd else 'Sacrifice Hydro'}.exe"
head = f"{title} {version}"
L = []
w = L.append
w(head); w("=" * len(head)); w("")
if hd:
    w("Hydro, god of water, joins the five gods of Sacrifice as a sixth god - together with the")
    w("high-resolution textures of Sacrifice HD, built from your own copy of the game.")
else:
    w("Hydro, god of water, joins the five gods of Sacrifice as a sixth god, with ten new creatures,")
    w("nine new spells and its own manalith.")
w("")
w("THIS MOD NEEDS THE ORIGINAL GAME. None of Sacrifice's files are included: the mod uses the files")
w("of your copy of Sacrifice (GOG or Steam version) and does not work without them. Your game files")
w("are never changed.")
w("")
w("Requirements")
w("- Sacrifice (GOG or Steam version, or an installed CD version with its \"data\" and \"maps\" folders).")
w("- Windows 10 or 11, 64-bit, and a graphics card with OpenGL 4.x.")
w("- Microsoft Visual C++ Redistributable 2015-2022 (x64). Most PCs have it; if not, the setup offers")
w("  the download: https://aka.ms/vs/17/release/vc_redist.x64.exe")
if hd:
    w("- About 300 MB of free space for the HD texture cache. Building it uses your graphics card")
    w("  through Vulkan (Real-ESRGAN); without Vulkan it uses a slower built-in upscaler.")
w("")
w("Install")
w("- Unzip the Setup download (right-click, \"Extract All...\") and run")
w(f"  \"{setup}\" in the unzipped folder.")
w("  It finds your Sacrifice folder (or lets you pick it), checks that the original game is there")
w(f"  and installs the mod into Sacrifice\\{folder}. It also makes Desktop and Start menu")
w("  shortcuts and an entry in Windows' Apps & features to uninstall it.")
w("- Updating: run the newer setup over the old one. Your settings, key bindings and saved games")
w("  are kept.")
w(f"- Manual install (instead of the setup): unzip the \"{folder}\" folder into your Sacrifice")
w(f"  folder, so that you get ...\\Sacrifice\\{folder}\\Sacrifice Hydro Launcher.exe next to")
w("  Sacrifice's \"data\" and \"maps\" folders. To remove it, delete that folder.")
w("- Uninstall: Apps & features, the Start menu, or")
w(f"  \"Uninstall {'Sacrifice Hydro HD' if hd else 'Sacrifice Hydro'}.exe\" in the mod's folder. It asks whether to keep")
w("  your settings and saved games.")
if hd:
    w("- The HD textures are built the first time you press Play (a few minutes, once). They are kept in")
    w("  %LOCALAPPDATA%\\SacrificeHD\\cache, which Sacrifice HD uses too; delete that folder to free the")
    w("  space if you remove both.")
w("")
w("Play")
w(f"- Start \"{'Sacrifice Hydro HD' if hd else 'Sacrifice Hydro'}\" from the Desktop or Start menu (Sacrifice Hydro Launcher.exe).")
w("  The launcher is the game's start menu:")
w("    Single player   you against 1 to 5 computer wizards, each with the god you choose (or Random),")
w("                    on any skirmish map, including maps you add to Sacrifice's maps folder.")
w("    Host a game     a multiplayer game; the others join with your IP address.")
w("    Join a game     enter the host's address. Everyone needs the same version of the mod.")
w("  Choose your name, god (Hydro, the five gods of Sacrifice, or Random) and wizard, then Play.")
w("- Game settings... - the original Game Settings: game type (Skirmish, Slaughter, Soul Harvest,")
w("  Domination), the rules (Random Altar, Observers, Custom Spellbooks, Allied Vision, Friendly Fire,")
w("  Always Gib, No Map Creatures/Souls, ...) and the settings (starting souls, levels, AI damage,")
w("  gib chance, XP rate, friendly damage).")
w("- Spellbook... - your own spellbook (the original game unlocked it after the campaign).")
w("- Options... - screen, sound and mouse settings; the Picture page (colours - vivid like the")
w("  original game or filmic -, brightness, colour strength, contrast, texture sharpness, VSync and")
w("  a frame rate limit); and the Key bindings page.")
w("- Load game... - continue a saved game.")
w("- In a game: Esc opens the game menu (Save Game, Load Game, Surrender, Resume Game); F1 is a quick")
w("  save and F2 a quick load. Saved games are in the mod's \"saves\" folder (single player only).")
w("  When every enemy is defeated the Victory window appears; OK closes the game.")
w("")
w("Hydro")
w("Creatures (level, souls, health, attack, ability):")
rows = [
    ("Tidekin", 1, 1, 800, "melee 160", "Shell Up - its shell halves damage taken"),
    ("Reefstalker", 1, 1, 545, "Water Bow", "a naga archer"),
    ("Mistwing", 1, 1, 450, "Brine Spray", "Bubble Shield - shields an ally; flies"),
    ("Mantis Shrimp", 2, 2, 1235, "Slowing Bolt", "long range; slows what it hits"),
    ("Brineclaw", 3, 2, 1800, "melee 280", "Crushing Grip - drags an enemy into its claws"),
    ("Tidecaller", 4, 2, 1300, "Water Bolt 250", "Tidal Mend - heals friendly creatures nearby"),
    ("Puffy", 5, 3, 1900, "Water Bolt 400", "Toxic Spines - poisons every enemy around it"),
    ("Reefbreaker", 6, 3, 1450, "Whirlpool Mortar 500", "a hermit-crab cannon"),
    ("Abyss Guardian", 7, 4, 4500, "melee 700", "Tidal Slam - knocks enemies down around it"),
    ("Leviathan", 9, 5, 5200, "melee 800", "Tidal Breath (water blast), Maelstrom"),
]
for n, lv, s, hp, atk, ab in rows:
    w(f"  {n:<15}{lv}  {s}  {hp:>4}  {atk:<21}{ab}")
w("Spells:")
spells = [
    ("1", "Water Jet", "a jet of water: Fireball damage and a knock-back"),
    ("2", "Bubble Armor", "a bubble that cuts damage and heals the wizard"),
    ("3", "Undertow", "a whirlpool that snares a creature and slows those around it"),
    ("4", "Healing Spring", "a spring that heals your side standing in it"),
    ("5", "Geyser", "a column of water that damages and throws enemies"),
    ("6", "Tidal Wall", "a waterfall wall that stops missiles and slows enemies"),
    ("7", "Acid Rain", "corrosive rain that eats through protections"),
    ("8", "Tsunami", "a huge wave that sweeps enemies away"),
    ("8", "Kraken", "a giant whirlpool with tentacles that seize anything - friend or foe"),
]
for lv, n, d in spells:
    w(f"  {lv}  {n:<16}{d}")
w("- Hydro has its own manalith (coral-crowned reef rock), sounds, voices and spell effects.")
w("- The new creatures are built on the skeletons of the original ones and move with their")
w("  animations.")
w("")
w("Also new in the engine")
w("- The original Game Settings, custom spellbooks, the Victory and Defeat windows, the in-game")
w("  menu with saving, loading and surrender, and a key binding editor.")
if hd:
    w("- The HD textures of Sacrifice HD (upscaled with Real-ESRGAN from your copy of the game).")
    w("- If you also use the separate Sacrifice HD mod, both share the texture cache; keep both at")
    w("  the versions released together so the cache is not rebuilt.")
w("")
w("Known limits")
w("- Mission Objectives, Speech History and the in-game Options are not available in skirmish")
w("  games; the campaign is not part of the mod. View Summary after a defeat is not available yet.")
w("- Saving and loading work in single player games.")
w("- Multiplayer: every player needs the same version of the mod.")
w("")
w("Files")
w(f"- Sacrifice Hydro Launcher.exe - start here. Hydro.exe - the game (sacengine with the Hydro mod).")
w("- mods\\hydro - Hydro's creatures, spells, manalith, effects and sounds.")
w("- settings.txt - the options (the launcher edits it). hotkeys.txt - the key bindings.")
w("- source\\SacrificeHydro-source.zip - the complete source code (GPL-3.0).")
if hd:
    w("- HDTextures - the Sacrifice HD texture builder.")
w("")
w("Credits and licences")
w("- Sacrifice by Shiny Entertainment (Interplay). This is an unofficial fan mod; it includes none of")
w("  the game's files.")
w("- sacengine by tg and contributors (https://github.com/tg-2/sacengine), GPL-3.0. Hydro.exe is a")
w("  modified version and is therefore also GPL-3.0; its full source is in the source folder.")
w("- dagon and dlib by Timur Gafarov et al. (Boost), SDL2 (zlib), FreeType (FreeType licence),")
w("  OpenAL Soft (LGPL, unmodified DLL).")
w("- The launcher and setup use walk and win by The Walk Authors (BSD-3-Clause), go-ole (MIT),")
w("  golang.org/x/sys")
w("  (BSD-3-Clause) and govaluate (MIT).")
w("- Creature sounds made from CC0 packs: \"80 CC0 creature SFX\" and \"80 CC0 creature SFX #2\" by")
w("  rubberduck and \"Monster SFX pack\" by Ogrebane (opengameart.org) - with thanks.")
w("- Spoken lines synthesised with Piper by Michael Hansen / Rhasspy (MIT) using the")
w("  \"en-us-libritts-high\" voice, trained on LibriTTS by Heiga Zen et al. (CC BY 4.0,")
w("  http://www.openslr.org/60/), from LibriVox public-domain audiobooks.")
if hd:
    w("- Texture upscaling: Real-ESRGAN / realesrgan-ncnn-vulkan by Xintao Wang et al.")
    w("  (BSD-3-Clause / MIT).")
w("- Licence texts are in the licenses folder" + (" and HDTextures\\licenses." if hd else "."))
w("")
open(out, "w", newline="\r\n").write("\n".join(L) + "\n")
