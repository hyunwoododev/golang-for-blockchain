package db

import (
	"github.com/boltdb/bolt"
	"github.com/hyunwoododev/golang-for-blockchain/utils"
)

const (
	dbName       = "blockchain.db" // 데이터베이스 파일 이름
	dataBucket   = "data"          // 체크포인트와 관련된 데이터가 저장될 버킷 이름
	blocksBucket = "blocks"        // 블록 데이터를 저장할 버킷 이름
	checkpoint   = "checkpoint"    // 블록체인의 마지막 상태를 저장하는 키
)

var db *bolt.DB

func DB() *bolt.DB {
	if db == nil {
		dbPointer, err := bolt.Open(dbName, 0600, nil)
		db = dbPointer
		utils.HandleErr(err)
		
		// 데이터베이스 트랜잭션을 통해 필요한 버킷이 없으면 생성
		err = db.Update(func(t *bolt.Tx) error {
			// dataBucket이 존재하지 않으면 생성
			_, err := t.CreateBucketIfNotExists([]byte(dataBucket))
			utils.HandleErr(err)
			
			// blocksBucket이 존재하지 않으면 생성
			_, err = t.CreateBucketIfNotExists([]byte(blocksBucket))
			return err
		})
		utils.HandleErr(err)
	}
	return db
}


func Close() {
	DB().Close()
}

func SaveBlock(hash string, data []byte) {
	err := DB().Update(func(t *bolt.Tx) error {
		bucket := t.Bucket([]byte(blocksBucket))
		err := bucket.Put([]byte(hash), data)
		return err
	})
	utils.HandleErr(err)
}

// SaveCheckpoint 함수는 현재 블록체인의 체크포인트 데이터를 저장한다.
// 체크포인트는 마지막 블록 또는 체인의 상태를 나타낸다.
func SaveCheckpoint(data []byte) {
	err := DB().Update(func(t *bolt.Tx) error {
		bucket := t.Bucket([]byte(dataBucket))    // dataBucket에 접근
		err := bucket.Put([]byte(checkpoint), data) // checkpoint 키로 데이터 저장
		return err
	})
	utils.HandleErr(err) // 오류 발생 시 처리
}

// Checkpoint 함수는 저장된 블록체인의 체크포인트 데이터를 반환한다.
func Checkpoint() []byte {
	var data []byte
	DB().View(func(t *bolt.Tx) error {           // 읽기 전용 트랜잭션 사용
		bucket := t.Bucket([]byte(dataBucket))    // dataBucket에 접근
		data = bucket.Get([]byte(checkpoint))     // checkpoint 키에 해당하는 데이터 조회
		return nil
	})
	return data // 조회한 데이터 반환
}

// Block 함수는 주어진 해시 값에 해당하는 블록 데이터를 반환한다.
func Block(hash string) []byte {
	var data []byte
	DB().View(func(t *bolt.Tx) error {          // 읽기 전용 트랜잭션 사용
		bucket := t.Bucket([]byte(blocksBucket)) // blocksBucket에 접근
		data = bucket.Get([]byte(hash))          // 해시 키에 해당하는 블록 데이터 조회
		return nil
	})
	return data // 조회한 블록 데이터 반환
}
