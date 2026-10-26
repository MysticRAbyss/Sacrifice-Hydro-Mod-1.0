Sacrifice: Hydro 1.0.0 - source code
====================================

Sacrifice: Hydro adds a sixth god, Hydro, with 10 creatures and 9 spells, to Sacrifice (Shiny
Entertainment, 2000). It runs on sacengine, tg's open-source re-implementation of the game's
engine, which reads the player's own copy of Sacrifice. Sacrifice Hydro HD is the same mod with
the HD textures of Sacrifice HD. None of Sacrifice's own files are part of the mod or of this
source.

This is the complete source of both downloads, as the GPL-3.0 licence of sacengine requires.
BUILD.txt explains step by step how to build everything, and how to check your build against
the published 1.0.0 (CHECKSUMS.txt).

What is where
-------------
engine\                       sacengine (https://github.com/tg-2/sacengine) by tg and contributors,
                              with the Hydro changes (GPL-3.0); it builds Hydro.exe. The Hydro code is
                              mostly in source\hydromod.d and source\hydrosave.d and is marked
                              "Hydro mod" elsewhere.
  dagon\                        tg's fork of the dagon 3D engine (Boost licence), with Hydro's picture
                                options, and its DerelictGL3 / DerelictSDL2 bindings
  dlib\                         dlib (Boost licence)
  .git\HEAD, .git\refs\         only the commit name (6b0082f) - the engine shows it in its version
                                and multiplayer games use it to check that both players run the same
hydro-engine-changes.patch    all Hydro changes to sacengine as one patch, against commit 181d1ce
hydro-dagon-changes.patch     all Hydro changes to dagon, against commit 3ccbb7a
launcher\                     Sacrifice Hydro Launcher.exe (Go, lxn/walk): Play, Load game, options
                              with the Game / Picture / Key bindings tabs, the original-game check
installer\                    the setup / uninstaller (Go; lxn/walk, go-ole, golang.org/x/sys)
  editions\                     the settings of each download (hydro.json, hydrohd.json; sachd.json is
                                Sacrifice HD's)
  verinfo\                      writes the icon, manifest and version information into the programs
                                (uses goversioninfo, MIT, and akavel/rsrc, MIT)
  core_test.go                  tests of the install / update / uninstall logic ("go test .")
release\                      the tool that builds both downloads (see BUILD.txt)
package\                      the small files that go into the mod: settings.txt (hd\settings.txt for
                              Hydro HD), hotkeys.txt, the licences, and HD's prepare.bat
runtime\                      SDL2.dll, OpenAL32.dll and freetype.dll, the libraries Hydro.exe uses,
                              exactly as shipped (see "Third-party parts" below)
tools\
  kappagen\                     (Go) builds the Hydro creature models and textures from the original
                                creatures' skeletons; it reads your own Sacrifice data files
  sfx\                          (Python) the creature and spell icons, effect textures, the manalith
                                and the sounds
  release\make_readme.py        writes the README.txt of each download

The creature and spell files of the mod (the "mods" folder) are made with tools\kappagen and
tools\sfx from your own game files, so they are not in this folder as finished files; the release
tool takes them from the published Manual download instead (BUILD.txt, step 4).

Third-party parts
-----------------
- SDL2 2.30.9 (zlib licence): SDL2.dll from SDL2-2.30.9-win32-x64.zip,
  https://github.com/libsdl-org/SDL/releases/tag/release-2.30.9
- OpenAL Soft 1.24.1 (LGPL-2.1): bin\Win64\soft_oal.dll from openal-soft-1.24.1-bin.zip
  (https://openal-soft.org), renamed OpenAL32.dll
- FreeType 2.14.3 (FreeType licence): a 64-bit Windows build of freetype.dll
- The HD texture builder in Sacrifice Hydro HD is Sacrifice HD's (its own source is
  SacrificeHD-1.1.1-source.zip), with Real-ESRGAN (MIT / BSD-3-Clause)
- Go libraries: lxn/walk and lxn/win (BSD-3-Clause), go-ole (MIT), golang.org/x/sys (BSD-3-Clause),
  govaluate (MIT), akavel/rsrc and goversioninfo (MIT)
The licences of everything the mod ships are in package\licenses.
