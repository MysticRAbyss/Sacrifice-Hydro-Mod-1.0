// copyright © tg
// distributed under the terms of the gplv3 license
// https://www.gnu.org/licenses/gpl-3.0.txt

import util,assets;
import dlib.image,dlib.image.color;
import std.stdio, std.string, std.range, std.algorithm, std.path, std.exception;

SuperImage loadTXTR(string filename){
	int numFrames,frameWidth,frameHeight;
	return loadTXTR(filename,numFrames,frameWidth,frameHeight);
}

SuperImage loadTXTR(string filename,out int numFrames,out int frameWidth,out int frameHeight){
	filename=fixPath(filename);
	enforce(filename.endsWith(".TXTR")||filename.endsWith(".ICON"));
	auto base = filename[0..$-".TXTR".length];
	char[4] tag;
	import std.utf:byChar;
	copy(base[$-4..$].byChar.retro,tag[].byChar);
	if(tag in textureReplacementsSloppy) return textureReplacementsSloppy[tag];
	if(replacementPath(base) in textureReplacements) return textureReplacements[replacementPath(base)];
	auto txt=readFile(filename);
	auto width=parseLE(txt[0..4]);
	txt=txt[4..$];
	auto height=parseLE(txt[0..4]);
	txt=txt[4..$];
	auto tmp=txt[0..4].dup;
	txt=txt[4..$];
	reverse(tmp);
	bool grayscale=false;
	string pal=cast(string)tmp.idup;
	auto hasAlphaColor=!!(txt[0]&4);
	auto hasExplicitAlpha=!!(txt[2]&1);
	ubyte[3] alphaColor=[txt[6],txt[5],txt[4]];
	txt=txt[8..$];
	numFrames=cast(int)parseLE(txt[0..4]);
	frameWidth=cast(int)parseLE(txt[4..8]);
	frameHeight=cast(int)parseLE(txt[8..12]);
	txt=txt[16..$]; // TODO: what is txt[12..16] ?
	ubyte[] alphaChannel;
	if(hasExplicitAlpha){
		auto rest=txt.length-width*height;
		int sc=0;
		foreach(k;[2,4,8]) if(width%k==0&&height%k==0&&rest==(width/k)*(height/k)) sc=k;
		if(txt.length==2*width*height){
			alphaChannel=txt[$-width*height..$];
		}else if(sc){ // Hydro mod: the colours were upscaled (Sacrifice HD) but not the alpha plane: scale it up too (bilinear)
			auto sw=width/sc, sh=height/sc;
			auto small=txt[width*height..$];
			alphaChannel=new ubyte[](width*height);
			foreach(y;0..height) foreach(x;0..width){
				float fx=max(0.0f,(x+0.5f)/sc-0.5f), fy=max(0.0f,(y+0.5f)/sc-0.5f);
				int x0=min(cast(int)fx,cast(int)sw-1), y0=min(cast(int)fy,cast(int)sh-1);
				int x1=min(x0+1,cast(int)sw-1), y1=min(y0+1,cast(int)sh-1);
				float tx=fx-x0, ty=fy-y0;
				float a=small[y0*sw+x0]*(1-tx)*(1-ty)+small[y0*sw+x1]*tx*(1-ty)+small[y1*sw+x0]*(1-tx)*ty+small[y1*sw+x1]*tx*ty;
				alphaChannel[y*width+x]=cast(ubyte)min(255.0f,a+0.5f);
			}
		}else{ // Hydro mod: a texture rescaled by another tool (e.g. Sacrifice HD) may keep a smaller or no alpha plane; drawn opaque then
			import std.process:environment; if(environment.get("HYDRO_DEBUG","").length) stderr.writeln("hydro mod: ",filename,": alpha plane ",txt.length-width*height," bytes for ",width,"x",height,"; ignored");
			hasExplicitAlpha=false;
		}
	}
	enforce(txt.length>=width*height,filename);
	txt=txt[0..width*height];
	ubyte[] palt;
	if(pal.toLower()=="gray"){
		grayscale=true;
	}else{
		auto palFile=buildPath(dirName(filename), pal~".PALT");
		palt=readFile(palFile);
		palt=palt[8..$]; // header bytes (TODO: figure out what they mean)
	}
	auto channels=3+(hasAlphaColor||hasExplicitAlpha);
	auto img=image(width, height, channels);
	auto data=img.data;
	static foreach(schannels;3..5){
		if(channels==schannels){
			foreach(i;0..data.length/schannels){
				auto ccol=txt[i];
				ubyte[3] tcol;
				if(grayscale) tcol=[ccol,ccol,ccol];
				else tcol=[palt[3*ccol+0],palt[3*ccol+1],palt[3*ccol+2]];
				data[schannels*i..schannels*i+3]=tcol[];
				static if(schannels==4){
					if(!(hasAlphaColor&&tcol==alphaColor))
						data[schannels*i+3]=alphaChannel.length?alphaChannel[i]:255;
				}
			}
			goto Lend;
		}
	}
Lend:;
	return img;
}
