package main

//Ce document regroupe quelques elements importants de language.

//------------------------------------------------------IMPORT----------------------------------------------------------
import (
	"fmt"
	"math/cmplx"
	"time"
)

//-----------------------------------------------ECRITURE DE FONCTION---------------------------------------------------
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

	//-----------------------------------------------------------VARIABLE-----------------------------------------------
	var declarationtType = 10
	fmt.Println("variable typee", declarationtType)
	var a, b, d = "vive", "le", "confinement"
	fmt.Println(a, b, d)

	//-------------------------------------------------TYPAGEDINAMIQUE--------------------------------------------------
	typage_dynamique := 10.5
	fmt.Println("typage dinamique", typage_dynamique)

	c, python, java := 2.2, false, "ouch"
	fmt.Println(c, python, java)

	//--------------------------------------------------REGROUPER LES TYPAGES-------------------------------------------
	var (
		ToBe          = false
		MaxInt uint64 = 1<<64 - 1
		z             = cmplx.Sqrt(-5 + 12i)
	)
	fmt.Printf("Type: %T Value: %v\n", ToBe, ToBe)
	fmt.Printf("Type: %T Value: %v\n", MaxInt, MaxInt)
	fmt.Printf("Type: %T Value: %v\n", z, z)

	//---------------------------------------------------Conversions de type--------------------------------------------
	var f = 12.2
	var j = uint(f)
	fmt.Print(j)

	//---------------------------------------------------CONSTANTE------------------------------------------------------
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

	//--------------------------------------------------FOR-------------------------------------------------------------

	//on ecrit une boucle for ainsi :
	sum := 0
	for i := 0; i < 10; i++ {
		sum += i
	}

	//--------------------------------------------------WHILE-----------------------------------------------------------

	// en Go le while s'ecrit avec un for.
	var masomme int = 0
	for masomme < 1000 {
		masomme = masomme + 250
	}
	fmt.Println(masomme)
	//---------------------------------------------------IF et ELSE-----------------------------------------------------
	variable := 22
	if variable < 0 {
		//action ici
	}

	// on peut aussi faire quelque actions avant le if, attention la variable v est donc locale.
	/*
		if v := math.Pow(x, n); v < lim {
			return v
		}

		if v := math.Pow(x, n); v < lim {
			return v
		} else {
			fmt.Printf("%g >= %g\n", v, lim)
		}
	*/

	//---------------------------------------------------SWITCH---------------------------------------------------------
	fmt.Println("When's Saturday?")
	today := time.Now().Weekday()
	fmt.Println(today)
	switch time.Saturday {
	case today + 0:
		fmt.Println("Today.")
	case today + 1:
		fmt.Println("Tomorrow.")
	case today + 2:
		fmt.Println("In two days.")
	default:
		fmt.Println("Too far away.")
	}

	// ou bien :
	t := time.Now()
	switch {
	case t.Hour() < 12:
		fmt.Println("Good morning!")
	case t.Hour() < 17:
		fmt.Println("Good afternoon.")
	default:
		fmt.Println("Good evening.")
	}

	//
	//---------------------------------------------------DEFER----------------------------------------------------------

	// ça peut servir, qui sait.
	defer fmt.Println("world")
	fmt.Println("hello")
}
