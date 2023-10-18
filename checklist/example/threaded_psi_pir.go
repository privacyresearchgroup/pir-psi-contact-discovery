package main

import (
	"math/rand"
	"math"
	"fmt"
	"sync"
	"time"
	pir_psi "checklist/pir-psi"
	db "checklist/db"
)

func rangeIn(low, hi int) uint64 {
    return uint64(low + rand.Intn(hi-low))
}

func main() {

	contacts := make([]uint64, 1000)
	for i:=0; i < len(contacts); i++ {
		contacts[i] = rangeIn(1000000000, 9999999999)
	}

	nBuckets := int(math.Pow(2, 26))
	nPartitions := 64
	nRows := int(math.Pow(2, 27))
	dummyQueries := 84

	metaSize := uint8(0)

	data := db.New(nBuckets, nPartitions, nRows, metaSize)


	schemes := make([]*pir_psi.PIR_PSI, 10)

	for i:=0; i < 10; i++ {
		schemes[i] = pir_psi.New(data)
	}

	var wg sync.WaitGroup
	wg.Add(10)

	var lock sync.Mutex

	s1 := time.Now()
	for i:=0; i < 10; i++ {
		go func(i int) {
			defer wg.Done()
			schemes[i].OnlinePhaseThreaded(contacts, dummyQueries, 4, metaSize, &lock)
		} (i)
	}

	wg.Wait()


	t1 := time.Now()
	fmt.Print("Time to answer all client queries: ")
	fmt.Println(t1.Sub(s1))



}
