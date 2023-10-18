package main

import (
	"math/rand"
	"math"
	"fmt"
	pir_psi "checklist/pir-psi"
	db "checklist/db"
)

func rangeIn(low, hi int) uint64 {
    return uint64(low + rand.Intn(hi-low))
}

func main() {


	// dummies[i] should be number of dummy queries necessary for partitions[i] partitions
	partitions := []int{16, 32, 64}
	dummies := []int{214, 131, 84}


	// create 1000 random contacts
	contacts := make([]uint64, 1000)
	for i:=0; i < len(contacts); i++ {
		contacts[i] = rangeIn(1000000000, 9999999999)
	}

	// for each partition 
	for j:=0; j<len(partitions); j++ {

		fmt.Print(partitions[j])
		fmt.Println(" total partitions")


		nBuckets := int(math.Pow(2, 26))
		nPartitions := partitions[j]
		nRows := int(math.Pow(2, 27))
		dummyQueries := dummies[j]
		// number of bytes of account identifier/key data, set to 0 for none
		metadataLength := uint8(56)

		// create a CF database with above params
		data := db.New(nBuckets, nPartitions, nRows, metadataLength)

		// generates hints (this function prints offline metrics)
		scheme := pir_psi.New(data)

		// run online phase (this function prints online metrics)
		scheme.OnlinePhase(contacts, dummyQueries, 4, metadataLength)

	}



}
