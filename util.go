package main

import ()

var gPosition_attr uint32 = 1 // prog.GetAttribLocation("position")
var gUVs_attr uint32 = 2      // prog.GetAttribLocation("uvs")

func StringToArray(s string) []uint8 {
	var srcArray []uint8 = make([]uint8, len(s)+1)
	for i := 0; i < len(s); i++ {
		srcArray[i] = s[i]
	}
	srcArray[len(s)] = 0
	return srcArray
}
