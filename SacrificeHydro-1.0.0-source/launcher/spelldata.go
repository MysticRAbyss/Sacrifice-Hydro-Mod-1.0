package main

// generated from the game data (engine dump HYDRO_SPELLS=1): every god's creatures and spells in
// spellbook order, as (tag, name)

var godCreatures = map[string][11][2]string{
	"Persephone": {{"maho", "Manahoar"}, {"humn", "Druid"}, {"huma", "Ranger"}, {"hsbt", "Shrike"}, {"arac", "Scarab"}, {"trol", "Troll"}, {"lflp", "Gnome"}, {"sprl", "Gremlin"}, {"babc", "Mutant"}, {"lstm", "Ent"}, {"gdrg", "Dragon"}},
	"Pyro": {{"maho", "Manahoar"}, {"fdiz", "Cog"}, {"klpf", "Flame Minion"}, {"fhos", "Spitfire"}, {"farc", "Tickferno"}, {"ftrl", "Firefist"}, {"fflp", "Pyromaniac"}, {"fcak", "Pyrodactyl"}, {"fclw", "Bombard"}, {"fgun", "Warmonger"}, {"rdrg", "Phoenix"}},
	"James": {{"maho", "Manahoar"}, {"orcy", "Trogg"}, {"klap", "Earthfling"}, {"ehos", "Gargoyle"}, {"slug", "Basilisk"}, {"bear", "Taurock"}, {"eclw", "Flummox"}, {"ecak", "Ikarus"}, {"ball", "Boulderdash"}, {"estm", "Jabberocky"}, {"drag", "Rhinok"}},
	"Stratos": {{"maho", "Manahoar"}, {"gobl", "Frostwolf"}, {"rcha", "Sylph"}, {"sbob", "Brainiac"}, {"aarc", "Vortick"}, {"bala", "Squall"}, {"sogr", "Storm Giant"}, {"sprg", "Seraph"}, {"aclw", "Flurry"}, {"prcy", "Yeti"}, {"bdrg", "Silverback"}},
	"Charnel": {{"maho", "Manahoar"}, {"dizd", "Scythe"}, {"klpd", "Fallen"}, {"bzap", "Locust"}, {"dslg", "Necryl"}, {"dcak", "Blight"}, {"dprc", "Netherfiend"}, {"dflp", "Deadeye"}, {"gutc", "Abomination"}, {"dgun", "Styx"}, {"dmon", "Hellmouth"}},
	"Hydro": {{"maho", "Manahoar"}, {"tidk", "Tidekin"}, {"rfsk", "Reefstalker"}, {"mstw", "Mistwing"}, {"mshr", "Mantis Shrimp"}, {"brcl", "Brineclaw"}, {"tdcl", "Tidecaller"}, {"pufy", "Puffy"}, {"rfbr", "Reefbreaker"}, {"abgd", "Abyss Guardian"}, {"lvtn", "Leviathan"}},
}

var godSpells = map[string][11][2]string{
	"Persephone": {{"spup", "Speed up"}, {"heal", "Heal"}, {"flao", "Wrath"}, {"efrm", "Ethereal Form"}, {"vine", "Grasping Vines"}, {"rbow", "Rainbow"}, {"rfrg", "Rain of frogs"}, {"haur", "Healing Aura"}, {"vwal", "Vinewall"}, {"chrm", "Charm"}, {"retr", "Meanstalks"}},
	"Pyro": {{"spup", "Speed up"}, {"heal", "Heal"}, {"firb", "Fireball"}, {"ffrm", "Fireform"}, {"fpar", "Rings of fire"}, {"dfir", "Dragonfire"}, {"fexp", "Explosion"}, {"frwl", "Firewall"}, {"rfir", "Rain of fire"}, {"rage", "Blind Rage"}, {"volc", "Volcano"}},
	"James": {{"spup", "Speed up"}, {"heal", "Heal"}, {"rock", "Rock"}, {"skin", "Skin of stone"}, {"dcmp", "Soul Mole"}, {"erpt", "Erupt"}, {"halo", "Halo of Earth"}, {"swal", "Wall of spikes"}, {"metr", "Bombardment"}, {"bovi", "Bovine Intervention"}, {"bore", "Bore"}},
	"Stratos": {{"spup", "Speed up"}, {"heal", "Heal"}, {"litn", "Lightning"}, {"shld", "Air shield"}, {"frez", "Freeze"}, {"cltn", "Chain Lightning"}, {"swnd", "Soul Wind"}, {"froz", "Frozen Ground"}, {"fenc", "Fence"}, {"ckil", "Cloudkill"}, {"torn", "Tornado"}},
	"Charnel": {{"spup", "Speed up"}, {"heal", "Heal"}, {"swrm", "Insect Swarm"}, {"pswr", "Protective swarm"}, {"leth", "Slime"}, {"aded", "Animate Dead"}, {"rift", "Demonic Rift"}, {"wwal", "Wailing wall"}, {"plag", "Plague"}, {"ivap", "Intestinal Vaporization"}, {"deat", "Death"}},
	"Hydro": {{"spup", "Speed up"}, {"heal", "Heal"}, {"wjet", "Water Jet"}, {"bbar", "Bubble Armor"}, {"undt", "Undertow"}, {"hspr", "Healing Spring"}, {"gysr", "Geyser"}, {"tdwl", "Tidal Wall"}, {"acrn", "Acid Rain"}, {"tsnm", "Tsunami"}, {"krkn", "Kraken"}},
}
