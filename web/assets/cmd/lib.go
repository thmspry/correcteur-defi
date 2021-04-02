package main

import (
	"gitlab.univ-nantes.fr/E192543L/projet-s3/DAO"
	"syscall/js"
)

func getDefis(this js.Value, inputs []js.Value) interface{} {

	var res []interface{}
	defis := DAO.GetDefis()

	for _, defi := range defis {
		res = append(res, []interface{}{defi.Num, defi.DateDebut.String(), defi.DateFin.String(), defi.Correcteur})
	}

	return res
}

func main() {
	c := make(chan int)
	js.Global().Set("getDefis", js.FuncOf(getDefis))
	<-c
}
