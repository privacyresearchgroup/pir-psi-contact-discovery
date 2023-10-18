package main

import (
	"bytes"
	"log"
	"math/rand"
	"checklist/pir"
	"fmt"
	"time"
	"encoding/binary"
	"math"
	"sync"
)

func rangeIn(low, hi int) uint64 {
    return uint64(low + rand.Intn(hi-low))
}

// Generate a database filled with random bytes
func getRows(nRows int, rowLen int) []pir.Row {
	rows := make([]pir.Row, nRows)
	for i := 0; i < nRows; i++ {
		rows[i] = make(pir.Row, rowLen)

		phone_num := rangeIn(1000000000, 9999999999)
		bs := make([]byte, 8)
    	binary.LittleEndian.PutUint64(bs, phone_num)


    	copy(rows[i], bs)

	}

	return rows
}
func getDBandPartitions(rows []pir.Row, nPartitions int, nPir int) (*pir.StaticDB, []*pir.StaticDB) {
	db := pir.StaticDBFromRows(rows)

	partitions := make([]*pir.StaticDB, nPartitions)

	for i := 0; i < nPartitions; i++ {
		partitions[i] = pir.StaticDBFromRows(rows[i*nPir : i*nPir + nPir])
	}

	return db, partitions
}


func main() {

	pirType := pir.Punc


	nRows := int(math.Pow(2, 27))
	rowLen := 61
	rows := getRows(nRows, rowLen)

	dum := 84
	nPartitions := 64
	nPir := nRows/nPartitions


	db, partitions := getDBandPartitions(rows, nPartitions, nPir)

	offlineReq := pir.NewHintReq(pir.RandSource(), pirType)

	all_clients := make([][]pir.Client, 10)

	for j := 0; j < 10; j++ {
		clients := make([]pir.Client, nPartitions)
		for i := 0; i < nPartitions; i++ {
			offlineResp, err := offlineReq.Process(*partitions[i])
			if err != nil {
				log.Fatal("Offline hint generation failed")
			}

			clients[i] = offlineResp.(pir.HintResp).InitClient(pir.RandSource())
		}
		all_clients[j] = clients
	}


	var wg sync.WaitGroup
	wg.Add(10)
	var lock sync.Mutex

	s1 := time.Now()


	for cli := 0; cli < len(all_clients); cli++ {

		go func (cli int) {
			defer wg.Done()
			clients := all_clients[cli]
			for queryRow := 0; queryRow < 1000; queryRow++ {

				partition := int(math.Floor(float64(queryRow / nPir)))
				index := queryRow - (nPir*partition)


				//    Servers answer queries
				var err error

				lock.Lock()
				queries, recon := clients[partition].Query(index)
				lock.Unlock()

				answers := make([]interface{}, len(queries))
				for i := 0; i < len(queries); i++ {
					answers[i], err = queries[i].Process(*partitions[partition])
					if err != nil {
						log.Fatal("Error answering query")
					}
				}

				//    Client reconstructs
				row, err := recon(answers)
				if err != nil {
					log.Fatal("Could not reconstruct")
				}

				if !bytes.Equal(row, db.Row(queryRow)) {
					log.Fatal("Incorrect answer returned")
				}
			}

			var err error
			for part := 0; part < nPartitions; part++ {
				for j := 0; j < dum; j++ {

					lock.Lock()
					dummy := clients[part].DummyQuery()
					lock.Unlock()

					dummies := make([]interface{}, len(dummy))
					for i := 0; i < len(dummies); i++ {
						dummies[i], err = dummy[i].Process(*partitions[part])
						if err != nil {
							log.Fatal("Error answering query")
						}
					}

				}
			}
		} (cli)
	}

	wg.Wait()

	t1 := time.Now()
	fmt.Print("Time to answer all client queries: ")
	fmt.Println(t1.Sub(s1))


}
