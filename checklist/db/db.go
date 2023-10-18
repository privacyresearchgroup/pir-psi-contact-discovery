package db

import(
	cfilter "checklist/cfilter"
	"encoding/binary"
	"checklist/pir"
	"fmt"
	"time"
	"math/rand"
)


type DB struct {
	nBuckets  int // number of CF buckets
	NumPartitions int  // number of partitions
	nRows   int   // number of phone numbers in DB
	NumPir int // size of each partition
	Cf *cfilter.CFilter // CF holding data
	Partitions []*pir.StaticDB // partitioned CF
	BucketSize uint8
}

func New(nBuckets int, NumPartitions int, nRows int, mSize uint8) *DB {


	db := new(DB)

	db.nBuckets = nBuckets
	db.NumPartitions = NumPartitions
	db.nRows = nRows
	db.NumPir = nBuckets / NumPartitions

	fpSize := uint8(4)
	metaSize := uint8(mSize)

	db.BucketSize=3


	db.getCF(db.BucketSize, fpSize, metaSize)
	

	s2 := time.Now()
	db.getPartitions(db.getCFRows(fpSize, metaSize))
	t3 := time.Now()
	fmt.Print("Time to load get CF partitions: ")
	fmt.Println((t3.Sub(s2)))

	return db
}



func (db *DB) getCF(BucketSize uint8, fingerprintSize uint8, metadataSize uint8) {
	cf := cfilter.New(cfilter.Size(uint(db.nBuckets)), cfilter.BucketSize(BucketSize), cfilter.FingerprintSize(fingerprintSize), cfilter.MetadataSize(metadataSize))

	var phone_num uint64
	for i := 0; i < db.nRows; i++{

		phone_num = rangeIn(1000000000, 9999999999)
		
		bs := make([]byte, 8)
    	binary.LittleEndian.PutUint64(bs, phone_num)
		cf.Insert([]byte(bs))
	}

	db.Cf = cf
}

func (db *DB) getPartitions(rows []pir.Row)  {
	//db := pir.StaticDBFromRows(rows)

	partitions := make([]*pir.StaticDB, db.NumPartitions)


	for i := 0; i < db.NumPartitions; i++ {
		partitions[i] = pir.StaticDBFromRows(rows[i*db.NumPir : i*db.NumPir + db.NumPir])
	}

	db.Partitions = partitions
}




func(db *DB) getCFRows(fingerprintSize uint8, metadataSize uint8) []pir.Row{
	buckets := db.Cf.Buckets()

	totalLen := int(fingerprintSize + metadataSize)

	rows := make([]pir.Row, db.nBuckets)
	for i := 0; i < len(buckets); i++ {
		temp := make(pir.Row, int(db.BucketSize)*(totalLen))


		for j := 0; j < len(buckets[i]); j++{

			if len(db.Cf.FPKey(buckets[i][j])) > 0 {
				fpkey := db.Cf.FPKey(buckets[i][j])
				metadata := db.Cf.Metadata(buckets[i][j])
				for k := 0; k < len(fpkey); k++ {
					temp[j*totalLen + k] = fpkey[k]
				}
				for k := 0; k < len(metadata); k++ {
					temp[j*totalLen + (k + len(fpkey))] = metadata[k]
				}
			} else {
				for k:= 0; k < totalLen; k++ {
					temp[j*totalLen + k] = 0
				}
			}
		}
		rows[i] = temp
	}

	return rows
}

func rangeIn(low, hi int) uint64 {
    return uint64(low + rand.Intn(hi-low))
}

