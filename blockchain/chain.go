package blockchain

import (
	"sync"

	"github.com/hyunwoododev/golang-for-blockchain/db"
	"github.com/hyunwoododev/golang-for-blockchain/utils"
)

// 블록체인의 기본 난이도 및 조정 간격을 정의합니다.
const (
	defaultDifficulty  int = 2  // 기본 난이도
	difficultyInterval int = 5  // 난이도 조정 간격
	blockInterval      int = 2  // 블록 생성 목표 시간 (분 단위)
	allowedRange       int = 2  // 난이도 조정 허용 범위
)

// 블록체인 구조체 정의
type blockchain struct {
	NewestHash        string 	`json:"newestHash"`        // 가장 최근 블록의 해시
	Height            int    	`json:"height"`            // 블록체인의 높이 (블록 수)
	CurrentDifficulty int    	`json:"currentDifficulty"` // 현재 채굴 난이도
}

var b *blockchain
var once sync.Once // 싱글톤 패턴을 위한 sync.Once

// 블록체인 데이터를 복원하는 함수
func (b *blockchain) restore(data []byte) {
	utils.FromBytes(b, data) // 바이트 데이터를 블록체인 구조체로 변환
}

// 블록체인 상태를 저장하는 함수
func (b *blockchain) persist() {
	db.SaveCheckpoint(utils.ToBytes(b)) // 블록체인 구조체를 바이트로 변환하여 저장
}

// 새로운 블록을 추가하는 함수
func (b *blockchain) AddBlock(data string) {
	block := createBlock(data, b.NewestHash, b.Height+1) // 새로운 블록 생성
	b.NewestHash = block.Hash                            // 블록체인의 가장 최근 해시 업데이트
	b.Height = block.Height                              // 블록체인의 높이 업데이트
	b.CurrentDifficulty = block.Difficulty               // 블록체인의 난이도 업데이트
	b.persist()                                          // 블록체인 상태 저장
}

// 블록체인의 모든 블록을 반환하는 함수
func (b *blockchain) Blocks() []*Block {
	var blocks []*Block // 블록 포인터 슬라이스 선언
	hashCursor := b.NewestHash // 가장 최근 블록의 해시로 시작

	for {
		block, _ := FindBlock(hashCursor) // 현재 해시로 블록 찾기
		blocks = append(blocks, block)    // 찾은 블록을 슬라이스에 추가

		if block.PrevHash != "" { // 이전 블록의 해시가 존재하면
			hashCursor = block.PrevHash // 커서를 이전 블록의 해시로 이동
		} else {
			break // 제네시스 블록에 도달하면 루프 종료
		}
	}

	return blocks // 모든 블록을 포함한 슬라이스 반환
}

// 블록체인의 채굴 난이도를 재계산하는 함수
func (b *blockchain) recalculateDifficulty() int {
	allBlocks := b.Blocks() // 모든 블록 가져오기
	newestBlock := allBlocks[0] // 가장 최근 블록
	lastRecalculatedBlock := allBlocks[difficultyInterval-1] // 마지막으로 난이도 조정된 블록

	actualTime := (newestBlock.Timestamp / 60) - (lastRecalculatedBlock.Timestamp / 60) // 실제 생성 시간 계산
	expectedTime := difficultyInterval * blockInterval // 예상 생성 시간 계산

	if actualTime <= (expectedTime - allowedRange) { // 예상보다 빠르면 난이도 증가
		return b.CurrentDifficulty + 1
	} else if actualTime >= (expectedTime + allowedRange) { // 예상보다 느리면 난이도 감소
		return b.CurrentDifficulty - 1
	}

	return b.CurrentDifficulty // 예상 범위 내에 있으면 난이도 유지
}

// 블록체인의 채굴 난이도를 결정하는 함수
func (b *blockchain) difficulty() int {
	if b.Height == 0 { // 제네시스 블록일 경우
		return defaultDifficulty // 기본 난이도 반환
	} else if b.Height % difficultyInterval == 0 { // 난이도 조정 간격일 경우
		return b.recalculateDifficulty() // 난이도 재계산
	} else {
		return b.CurrentDifficulty // 현재 난이도 유지
	}
}

// 싱글톤 패턴으로 블록체인 인스턴스를 반환하는 함수
func Blockchain() *blockchain {
	if b == nil {
		once.Do(func() {
			b = &blockchain{
				Height: 0, // 초기 높이 설정
			}
			checkpoint := db.Checkpoint() // 체크포인트에서 데이터 로드
			if checkpoint == nil {
				b.AddBlock("Genesis") // 제네시스 블록 추가
			} else {
				b.restore(checkpoint) // 체크포인트 데이터로 복원
			}
		})
	}
	return b // 블록체인 인스턴스 반환
}