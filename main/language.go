package main

//Ce document regroupe les commandes de programmation basique relatif au language go.

//on import les paquets.
import (
	"fmt"
	"math/cmplx"
)

//exemple d'une fonction.
func add(x int, y int) int {
	return x + y
}

//ont peut aussi l'ecrire ainsi :

func add2(x, y, z int) int {
	return x + y + z
}

//elle peut retourner plusieurs valeur
func swap(x, y string) (string, string) {
	return y, x
}

//nommer les valeurs de retours permet de les utiliser dans la fonction, et de les retourner les indiquer.

func split(sum int) (x, y int) {
	x = sum * 4 / 9
	y = sum - x
	return
}

func needInt(x int) int { return x*10 + 1 }
func needFloat(x float64) float64 {
	return x * 0.1
}

func main() {
	//affichage des fonctions.
	fmt.Println("hello, World")
	fmt.Println("afficage fonction aadd", add(1, 2))
	fmt.Println(swap("za", "ae"))
	fmt.Println(split(17))

	//pour declarer des variables.
	var declarationt_type = 10
	fmt.Println("variable typee", declarationt_type)
	var a, b, d = "vive", "le", "confinement"
	fmt.Println(a, b, d)

	//typage dynamique (typage automatique)
	typage_dynamique := 10.5
	fmt.Println("typage dinamique", typage_dynamique)

	c, python, java := 2.2, false, "ouch"
	fmt.Println(c, python, java)

	// on peut choisir de regrouper ses declarations dans un bloc.
	var (
		ToBe          = false
		MaxInt uint64 = 1<<64 - 1
		z             = cmplx.Sqrt(-5 + 12i)
	)
	fmt.Printf("Type: %T Value: %v\n", ToBe, ToBe)
	fmt.Printf("Type: %T Value: %v\n", MaxInt, MaxInt)
	fmt.Printf("Type: %T Value: %v\n", z, z)

	//conversions de type
	var f = 12.2
	var j = uint(f)
	fmt.Print(j)

	// constante ( on ne peut la declarer avec la syntaxe :=)
	const world = "le monde va bien"

	//exemple de constantes numeriques ( je detaille pas c'est assez explicite)
	const (
		// Create a huge number by shifting a 1 bit left 100 places.
		// In other words, the binary number that is 1 followed by 100 zeroes.
		Big = 1 << 100
		// Shift it right again 99 places, so we end up with 1<<1, or 2.
		Small = Big >> 99
	)
	fmt.Println(needInt(Small))
	fmt.Println(needFloat(Small))
	fmt.Println(needFloat(Big))
	//fmt.Println(needInt(Big)) ca marche pas ce truc, normal.

}
