// Hydro mod: saved games
// A saved game is a recording of the game so far (as sacengine writes for --record) plus a key frame of
// the current state, so that --continue=<file> resumes it right away instead of re-simulating it.
// distributed under the terms of the gplv3 license
// https://www.gnu.org/licenses/gpl-3.0.txt
module hydrosave;

import std.algorithm, std.array, std.conv, std.file, std.path, std.string, std.datetime;
import recording_, state, util;

string savesDir(){
	import hydromod:modsRoot;
	return buildPath(modsRoot.length?modsRoot:getcwd(),"saves");
}

string cleanSaveName(string name){
	auto r=appender!string();
	foreach(dchar c;name.strip){
		if(c<32||"\\/:*?\"<>|".canFind(c)) continue;
		r.put(c);
	}
	auto s=r.data.strip;
	if(s.length>40) s=s[0..40].strip;
	return s.length?s:"Saved Game";
}

string saveFile(string name){ return buildPath(savesDir(),cleanSaveName(name)~".rcp"); }

struct SaveInfo{
	string name;
	string file;
	string[] details;
	SysTime time;
}

SaveInfo[] listSaves(){
	SaveInfo[] result;
	auto dir=savesDir();
	if(!exists(dir)) return result;
	foreach(entry;dirEntries(dir,"*.rcp",SpanMode.shallow)){
		SaveInfo s;
		s.file=entry.name;
		s.name=baseName(entry.name,".rcp");
		s.time=entry.timeLastModified;
		auto txt=setExtension(entry.name,".txt");
		if(exists(txt)) try s.details=readText(txt).lineSplitter.map!(l=>l.idup).filter!(l=>l.length>0).array; catch(Exception){}
		if(!s.details.length) s.details=["Saved: "~formatTime(s.time)];
		result~=s;
	}
	sort!((a,b)=>a.time>b.time)(result);
	return result;
}

string formatTime(SysTime t){
	return format("%04d-%02d-%02d %02d:%02d",t.year,cast(int)t.month,t.day,t.hour,t.minute);
}

// saves the committed state of the game and everything needed to continue from it
void saveGame(B)(GameState!B gameState,string name,string wizardName){
	auto dir=savesDir();
	mkdirRecurse(dir);
	auto file=saveFile(name);
	auto recording=new Recording!B(gameState,false);
	auto frame=gameState.committedFrame;
	Array!(Array!(Command!B)) commands;
	foreach(i;0..frame+1){
		if(i<gameState.commands.length) commands~=gameState.commands[i];
		else commands~=Array!(Command!B)();
	}
	recording.logCore=1;
	auto keyFrame=new ObjectState!B(recording.map,recording.sides,recording.proximity,recording.pathFinder,recording.triggers);
	keyFrame.copyFrom(gameState.committed);
	recording.core~=keyFrame;
	recording.finalize(commands);
	auto tmp=file~".tmp";
	recording.save(tmp);
	rename(tmp,file);
	auto mapName=baseName(gameState.current.map.path);
	mapName=mapName.stripExtension; // "(2) Trial.0123abcd" -> drop the version hash too
	if(mapName.length>9&&mapName[$-9]=='.') mapName=mapName[0..$-9];
	auto seconds=frame/60;
	string[] details;
	details~="Map: "~mapName;
	details~=format("Game time: %d:%02d:%02d",seconds/3600,seconds/60%60,seconds%60);
	details~="Saved: "~formatTime(Clock.currTime());
	if(wizardName.length) details~="Wizard: "~wizardName;
	std.file.write(setExtension(file,".txt"),details.join("\n")~"\n");
}

void deleteSave(string file){
	if(exists(file)) remove(file);
	auto txt=setExtension(file,".txt");
	if(exists(txt)) remove(txt);
}
