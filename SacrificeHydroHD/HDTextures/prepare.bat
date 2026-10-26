@echo off
rem Sacrifice Hydro HD: makes sure the high-resolution textures are built and current before the game
rem starts. The first time this takes a few minutes; afterwards it only checks them.
rem If the Sacrifice HD mod is installed and a game of it did not close properly, its HD files may
rem still be in the game's data folder: let it put the originals back first.
if exist "%~dp0..\..\SacrificeHD\swapped.json" if exist "%~dp0..\..\SacrificeHD\SacrificeHD.exe" "%~dp0..\..\SacrificeHD\SacrificeHD.exe" -restore
"%~dp0SacrificeHD.exe" -generate-only -game "%~dp0..\.."
