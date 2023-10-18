package PIR_PSI

import(
	"encoding/binary"
	"checklist/pir"
	"log"
	"fmt"
	"time"
    "math"
    "bytes"
    "sync"
    "strconv"
    db "checklist/db"
)


type PIR_PSI struct {
	db *db.DB // database of users
	clients []pir.Client // client for each partition
}




func New(db *db.DB) *PIR_PSI {


	scheme := new(PIR_PSI)

	scheme.db = db 

	scheme.GetHints()

	return scheme
}



func(scheme *PIR_PSI) GetHints() {

	count := 0

	s1 := time.Now()

	pirType := pir.Punc
	offlineReq := pir.NewHintReq(pir.RandSource(), pirType)

	clients := make([]pir.Client, scheme.db.NumPartitions)

	for i := 0; i < scheme.db.NumPartitions; i++ {

		offlineResp, err := offlineReq.Process(*scheme.db.Partitions[i])
		if err != nil {
			log.Fatal("Offline hint generation failed")
		}

		clients[i] = offlineResp.(pir.HintResp).InitClient(pir.RandSource())

		count = count + 1
	}

	scheme.clients = clients

	t2 := time.Now()
	fmt.Print("Hint generation: ")
	fmt.Println((t2.Sub(s1)))

	fmt.Print(count)
	fmt.Println(" total hints")
}

func(scheme *PIR_PSI) OnlinePhase(contacts []uint64, dummyQueries int, fingerprintSize uint8, metadataSize uint8) {


 	var currentIdx uint
    var err error
    equals := true

    clientCount := 0
    serverCount := 0

    totalLen := int(fingerprintSize + metadataSize)

    s1 := time.Now()

    // for each contact
    for num:= 0; num < len(contacts); num++ {
    	// get phone number hash
    	bs_check := getbs(contacts[num])
    	// find two possible bucket locations of phone number
    	a, b := scheme.db.Cf.ID(bs_check)

    	// for each location, PIR for the entire CF bucket
    	for i := 0; i < 2; i++ {
    		if i == 0 {
		        currentIdx = a
		    } else {
		        currentIdx = b
		    }

		    partition := int(math.Floor(float64(int(currentIdx) / scheme.db.NumPir)))
			index := int(currentIdx) - (scheme.db.NumPir*partition)

			// client queries for bucket index
    		queries, recon := scheme.clients[partition].Query(index)
    		clientCount = clientCount + 1

    		// server returns entire bucket
    		answers := make([]interface{}, len(queries))
			for k := 0; k < len(queries); k++ {
				answers[k], err = queries[k].Process(*scheme.db.Partitions[partition])
				if err != nil {
					log.Fatal("Error answering query")
				}

				serverCount = serverCount + 1
			}

			//  client reconstructs
			row, err := recon(answers)
			if err != nil {
				log.Fatal("Could not reconstruct")
			}

    		fp := scheme.db.Cf.FP(bs_check)

    		f := scheme.db.Cf.FPKey(fp)
    		//m := scheme.cf.Metadata(fp)

    		// check if phone number hash exists in bucket
    		equals = false
    		for j := 0; j < int(scheme.db.BucketSize); j++ {
    			foundFP := row[j*totalLen: j*totalLen + int(fingerprintSize)]
    			//foundMeta := row[j*totalLen + int(fingerprintSize): j*totalLen + totalLen]

    			if (bytes.Equal(f, foundFP)) {
    				equals = true
    				break
    			}
    		}

    		if (equals == true) {
    			fmt.Println("Match found!")
    			break
    		}
    	}

    	
    	if (equals == false) {
    		fmt.Println("Match not found!")
    	}

    }


    // dummy queries
    for part := 0; part < scheme.db.NumPartitions; part++ {
		for j := 0; j < dummyQueries; j++ {
			dummy := scheme.clients[part].DummyQuery()

			clientCount = clientCount + 1

			dummies := make([]interface{}, len(dummy))
			for i := 0; i < len(dummies); i++ {
				dummies[i], err = dummy[i].Process(*scheme.db.Partitions[part])
				if err != nil {
					log.Fatal("Error answering query")
				}

				serverCount = serverCount + 1
			}

		}
	}



	t3 := time.Now()
	fmt.Print("Online phase: ")
	fmt.Println((t3.Sub(s1)))

	fmt.Println("Number of client requests: " + (strconv.Itoa(clientCount)))
	fmt.Println("Number of server responses: " + (strconv.Itoa(serverCount)))



}

// same as above, but support multithreading
func(scheme *PIR_PSI) OnlinePhaseThreaded(contacts []uint64, dummyQueries int, fingerprintSize uint8, metadataSize uint8, lock *sync.Mutex) {

	//var lock sync.Mutex

 	var currentIdx uint
    var err error
    equals := true

    clientCount := 0
    serverCount := 0

    totalLen := int(fingerprintSize + metadataSize)

    //s1 := time.Now()


    for num:= 0; num < len(contacts); num++ {

    	bs_check := getbs(contacts[num])
    	a, b := scheme.db.Cf.ID(bs_check)

    	for i := 0; i < 2; i++ {
    		if i == 0 {
		        currentIdx = a
		    } else {
		        currentIdx = b
		    }

		    partition := int(math.Floor(float64(int(currentIdx) / scheme.db.NumPir)))
			index := int(currentIdx) - (scheme.db.NumPir*partition)

			lock.Lock()
    		queries, recon := scheme.clients[partition].Query(index)
    		lock.Unlock()

    		clientCount = clientCount + 1


    		answers := make([]interface{}, len(queries))
			for k := 0; k < len(queries); k++ {
				answers[k], err = queries[k].Process(*scheme.db.Partitions[partition])
				if err != nil {
					log.Fatal("Error answering query")
				}

				serverCount = serverCount + 1
			}

			//    Client reconstructs
			row, err := recon(answers)
			if err != nil {
				log.Fatal("Could not reconstruct")
			}

    		fp := scheme.db.Cf.FP(bs_check)

    		f := scheme.db.Cf.FPKey(fp)
    		//m := scheme.cf.Metadata(fp)


    		equals = false
    		for j := 0; j < int(scheme.db.BucketSize); j++ {
    			foundFP := row[j*totalLen: j*totalLen + int(fingerprintSize)]
    			//foundMeta := row[j*totalLen + int(fingerprintSize): j*totalLen + totalLen]

    			if (bytes.Equal(f, foundFP)) {
    				equals = true
    				break
    			}
    		}

    		if (equals == true) {
    			break
    		}
    	}
    }	

    // dummy queries
    for part := 0; part < scheme.db.NumPartitions; part++ {
		for j := 0; j < dummyQueries; j++ {

			lock.Lock() 

			dummy := scheme.clients[part].DummyQuery()

			lock.Unlock()

			clientCount = clientCount + 1

			dummies := make([]interface{}, len(dummy))
			for i := 0; i < len(dummies); i++ {
				dummies[i], err = dummy[i].Process(*scheme.db.Partitions[part])
				if err != nil {
					log.Fatal("Error answering query")
				}

				serverCount = serverCount + 1
			}

		}
	}

	
	/*
	t3 := time.Now()
	fmt.Print("Online phase: ")
	fmt.Println((t3.Sub(s1)))

	fmt.Print(clientCount)
	fmt.Println(" total client queries")

	fmt.Print(serverCount)
	fmt.Println(" total server answers")
	*/



}

func getbs(num uint64) []byte {
	bs := make([]byte, 8)
    binary.LittleEndian.PutUint64(bs, num)
    return bs
}

