package main

func origStrip(g *gameData, tag string, anims []string, yaw float64) {
	cr, err := loadOriginal(g, tag)
	if err != nil {
		panic(err)
	}
	frameStrip(cr, anims, 7, yaw, "/tmp/strip_"+tag+".png")
}

func dumpPitch(g *gameData, tag string, an string) {
	cr, _ := loadOriginal(g, tag)
	a := findAnim(cr, an)
	for i := 0; i < len(a.frames); i += len(a.frames) / 6 {
		pos, rot := boneTransforms(cr, &a.frames[i])
		back := qrot(rot[1], v3{0, -1, 0})
		println(an, i, int(pos[1][2]*100), int(back[0]*100), int(back[1]*100), int(back[2]*100))
	}
}
