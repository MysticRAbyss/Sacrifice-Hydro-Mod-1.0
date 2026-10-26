package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// compare part centroids of an engine OBJ export with our own skinning of the same frame
func compareOBJ(g *gameData, d *design, dir, objPath string) {
	cr, err := readBack(g, d, dir)
	if err != nil {
		panic(err)
	}
	a := findAnim(cr, d.bindAnim)
	if a == nil {
		a = findAnim(cr, "stance")
	}
	posed := pose(cr, &a.frames[0])
	f, _ := os.Open(objPath)
	sc := bufio.NewScanner(f)
	var parts [][]v3
	for sc.Scan() {
		l := sc.Text()
		if strings.HasPrefix(l, "o ") {
			parts = append(parts, nil)
		} else if strings.HasPrefix(l, "v ") {
			var x, y, z float64
			fmt.Sscanf(l, "v %f %f %f", &x, &y, &z)
			parts[len(parts)-1] = append(parts[len(parts)-1], v3{x, -z, y})
		}
	}
	cen := func(ps []v3) v3 {
		var c v3
		for _, p := range ps {
			c = c.add(p)
		}
		return c.mul(1 / float64(len(ps)))
	}
	for i := range posed {
		e := v3{}
		if i < len(parts) && len(parts[i]) > 0 {
			e = cen(parts[i])
		}
		fmt.Printf("part %d ours %.3v engine %.3v\n", i, cen(posed[i]), e)
	}
}
