package blockchain

import (
	"sync"

	"github.com/hyunwoododev/golang-for-blockchain/db"
	"github.com/hyunwoododev/golang-for-blockchain/utils"
)

// Difficulty adjustment constants
const (
	defaultDifficulty  int = 2 // 초기 블록 생성 시 기본 난이도
	difficultyInterval int = 5 // 난이도 재계산이 이루어지는 블록 간격
	blockInterval      int = 2 // 기대되는 블록 생성 간격(분 단위)
	allowedRange       int = 2 // 난이도 조정 시 허용되는 시간 오차 범위(분 단위)
)

// blockchain 구조체는 블록체인의 상태를 저장합니다.
type blockchain struct {
	NewestHash        string `json:"newestHash"`        // 가장 최근 블록의 해시
	Height            int    `json:"height"`            // 블록체인의 높이(블록 개수)
	CurrentDifficulty int    `json:"currentDifficulty"` // 현재 블록 생성 난이도
}

var b *blockchain        // 블록체인 인스턴스
var once sync.Once       // 싱글톤 패턴 구현을 위한 sync.Once

// restore 함수는 블록체인의 상태를 저장된 데이터로 복원합니다.
func (b *blockchain) restore(data []byte) {
	utils.FromBytes(b, data) // 직렬화된 데이터를 복원하여 blockchain 객체에 로드
}

// AddBlock 함수는 새 블록을 생성하고 블록체인의 상태를 갱신합니다.
func (b *blockchain) AddBlock() {
	block := createBlock(b.NewestHash, b.Height+1, getDifficulty(b)) // 새 블록 생성
	b.NewestHash = block.Hash                                       // 블록체인의 최신 해시 갱신
	b.Height = block.Height                                         // 블록체인의 높이 갱신
	b.CurrentDifficulty = block.Difficulty                         // 블록체인의 난이도 갱신
	persistBlockhain(b)                                             // 블록체인 상태 저장
}

// persistBlockhain 함수는 블록체인의 현재 상태를 저장합니다.
func persistBlockhain(b *blockchain) {
	db.SaveCheckpoint(utils.ToBytes(b)) // 블록체인의 상태를 직렬화하여 데이터베이스에 저장
}

// Blocks 함수는 블록체인의 모든 블록을 순서대로 반환합니다.
func Blocks(b *blockchain) []*Block {
	var blocks []*Block         // 블록을 저장할 슬라이스
	hashCursor := b.NewestHash  // 최신 해시에서 시작
	for {
		block, _ := FindBlock(hashCursor) // 현재 해시를 이용해 블록을 찾음
		blocks = append(blocks, block)   // 블록 슬라이스에 추가
		if block.PrevHash != "" {        // 이전 해시가 있는 경우 계속 탐색
			hashCursor = block.PrevHash
		} else { // 처음 블록(Genesis Block)까지 도달한 경우 중단
			break
		}
	}
	return blocks
}

// recalculateDifficulty 함수는 난이도를 재계산하여 반환합니다.
func recalculateDifficulty(b *blockchain) int {
	allBlocks := Blocks(b)                           // 모든 블록을 가져옴
	newestBlock := allBlocks[0]                      // 최신 블록
	lastRecalculatedBlock := allBlocks[difficultyInterval-1] // 마지막으로 난이도 조정된 블록
	actualTime := (newestBlock.Timestamp / 60) - (lastRecalculatedBlock.Timestamp / 60) // 실제 소요 시간(분 단위)
	expectedTime := difficultyInterval * blockInterval                                 // 기대되는 소요 시간(분 단위)

	// 난이도 조정
	if actualTime <= (expectedTime - allowedRange) { // 예상 시간보다 빠르면 난이도 증가
		return b.CurrentDifficulty + 1
	} else if actualTime >= (expectedTime + allowedRange) { // 예상 시간보다 느리면 난이도 감소
		return b.CurrentDifficulty - 1
	}
	return b.CurrentDifficulty // 그렇지 않으면 현재 난이도를 유지
}

// getDifficulty 함수는 현재 블록 생성 난이도를 반환합니다.
func getDifficulty(b *blockchain) int {
	if b.Height == 0 { // 최초 블록(Genesis Block)인 경우 기본 난이도 반환
		return defaultDifficulty
	} else if b.Height%difficultyInterval == 0 { // 난이도 재계산 시점인 경우
		return recalculateDifficulty(b)
	} else { // 그 외에는 현재 난이도를 유지
		return b.CurrentDifficulty
	}
}

// UTxOutsByAddress 함수는 특정 주소의 UTXO(사용되지 않은 트랜잭션 출력)를 반환합니다.
func UTxOutsByAddress(address string, b *blockchain) []*UTxOut {
	var uTxOuts []*UTxOut                      // 반환할 UTXO 슬라이스
	creatorTxs := make(map[string]bool)        // 이미 사용된 트랜잭션 ID를 저장할 맵

	for _, block := range Blocks(b) {          // 모든 블록을 순회
		for _, tx := range block.Transactions { // 각 블록의 트랜잭션을 확인
			for _, input := range tx.TxIns {   // 트랜잭션 입력 확인
				if input.Owner == address {    // 주소가 소유자인 경우
					creatorTxs[input.TxID] = true // 해당 트랜잭션 ID를 기록
				}
			}
			for index, output := range tx.TxOuts { // 트랜잭션 출력 확인
				if output.Owner == address {      // 주소가 소유자인 출력
					if _, ok := creatorTxs[tx.ID]; !ok { // 사용되지 않은 출력인 경우
						uTxOut := &UTxOut{tx.ID, index, output.Amount} // UTXO 생성
						if !isOnMempool(uTxOut) {                      // 메모리풀에 없는 경우
							uTxOuts = append(uTxOuts, uTxOut)          // 결과에 추가
						}
					}
				}
			}
		}
	}
	return uTxOuts
}

// BalanceByAddress 함수는 특정 주소의 잔액을 계산하여 반환합니다.
func BalanceByAddress(address string, b *blockchain) int {
	txOuts := UTxOutsByAddress(address, b) // 해당 주소의 모든 UTXO 가져오기
	var amount int
	for _, txOut := range txOuts { // 모든 UTXO의 금액을 합산
		amount += txOut.Amount
	}
	return amount
}

// Blockchain 함수는 싱글톤 패턴으로 블록체인 인스턴스를 반환합니다.
func Blockchain() *blockchain {
	once.Do(func() { // 한번만 실행
		b = &blockchain{Height: 0}        // 초기 블록체인 생성
		checkpoint := db.Checkpoint()    // 저장된 상태 확인
		if checkpoint == nil {           // 저장된 상태가 없으면
			b.AddBlock()                 // 새 블록 추가
		} else {                         // 저장된 상태가 있으면 복원
			b.restore(checkpoint)
		}
	})
	return b // 블록체인 인스턴스 반환
}
