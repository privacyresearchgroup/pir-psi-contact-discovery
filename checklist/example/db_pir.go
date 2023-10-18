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
	"strconv"
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

// partition DB
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

	// creates DB with nRows rows of rowLen byte records
	nRows := int(math.Pow(2, 27))
	rowLen := 61
	rows := getRows(nRows, rowLen)


	// number of partitions and dummy queries (will be iterated through)
	// dummies[i] should be number of dummy queries necessary for ps[i] partitions
	ps := []int{16, 32, 64}
	dummies := []int{214, 131, 84}

	for q:=0; q<len(ps); q++ {

		dum := dummies[q]
		nPartitions := ps[q]
		nPir := nRows/nPartitions

		fmt.Print(nPartitions)
		fmt.Println(" partitions")

		// create partitioned DB with above parameters
		db, partitions := getDBandPartitions(rows, nPartitions, nPir)


		s1 := time.Now()

		// Generate and store hints for each partitions
		offlineReq := pir.NewHintReq(pir.RandSource(), pirType)

		clients := make([]pir.Client, nPartitions)

		for i := 0; i < nPartitions; i++ {
			offlineResp, err := offlineReq.Process(*partitions[i])
			if err != nil {
				log.Fatal("Offline hint generation failed")
			}

			clients[i] = offlineResp.(pir.HintResp).InitClient(pir.RandSource())
		}

		t2 := time.Now()
		fmt.Print("Offline Phase: ")
		fmt.Println((t2.Sub(s1)))



		totalClient := 0
		totalServer := 0

		s2 := time.Now()
		// query 1000 DB rows
		for queryRow := 0; queryRow < 1000; queryRow++ {
			// find partition and physical location of index
			partition := int(math.Floor(float64(queryRow / nPir)))
			index := queryRow - (nPir*partition)


			// server answer queries
			var err error

			// client query for index
			queries, recon := clients[partition].Query(index)
			totalClient = totalClient + 1

			// server answers query
			answers := make([]interface{}, len(queries))
			for i := 0; i < len(queries); i++ {
				answers[i], err = queries[i].Process(*partitions[partition])
				if err != nil {
					log.Fatal("Error answering query")
				}
				totalServer = totalServer + 1
			}

			// client reconstructs
			row, err := recon(answers)
			if err != nil {
				log.Fatal("Could not reconstruct")
			}

			if !bytes.Equal(row, db.Row(queryRow)) {
				log.Fatal("Incorrect answer returned")
			}
		}


		// send dummy queries
		var err error
		for part := 0; part < nPartitions; part++ {
			for j := 0; j < dum; j++ {
				dummy := clients[part].DummyQuery()
				totalClient = totalClient + 1

				dummies := make([]interface{}, len(dummy))
				for i := 0; i < len(dummies); i++ {
					dummies[i], err = dummy[i].Process(*partitions[part])
					if err != nil {
						log.Fatal("Error answering query")
					}
					totalServer = totalServer + 1
				}

			}
		}

		t3 := time.Now()
		fmt.Print("Online phase: ")
		fmt.Println((t3.Sub(s2)))



		fmt.Println("Number of client requests: " + (strconv.Itoa(totalClient)))
		fmt.Println("Number of server responses: " + (strconv.Itoa(totalServer)))
	}


}
