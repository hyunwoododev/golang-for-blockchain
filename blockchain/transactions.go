package blockchain

import (
	"errors"
	"time"

	"github.com/hyunwoododev/golang-for-blockchain/utils"
)

// 채굴자가 블록을 생성할 때 받을 보상 금액
const (
	minerReward int = 50 // 정해진 채굴 보상. 채굴자는 새로운 블록을 추가하면 이 보상을 받음.
)

// Mempool은 트랜잭션을 대기시키는 공간. 새로운 트랜잭션이 블록에 포함되기 전까지 Mempool에 저장됨.
type mempool struct {
	Txs []*Tx // 대기 중인 트랜잭션들의 목록
}

// 전역으로 사용할 Mempool 변수. 모든 노드에서 동일한 Mempool을 공유한다고 가정.
var Mempool *mempool = &mempool{}

// Tx는 블록체인의 트랜잭션을 나타냄.
type Tx struct {
	ID        string   `json:"id"`        // 트랜잭션의 고유 식별 ID. 내용 기반 해시로 생성.
	Timestamp int      `json:"timestamp"` // 트랜잭션이 생성된 Unix 타임스탬프.
	TxIns     []*TxIn  `json:"txIns"`     // 트랜잭션의 입력 목록. (어디에서 돈을 가져왔는지)
	TxOuts    []*TxOut `json:"txOuts"`    // 트랜잭션의 출력 목록. (어디로 돈을 보낼 것인지)
}

// 트랜잭션의 ID를 생성하는 함수
func (t *Tx) getId() {
	// 트랜잭션 데이터를 모두 합쳐서 고유한 해시 값을 만듦. 이 값이 트랜잭션의 ID가 됨.
	t.ID = utils.Hash(t)
}

// TxIn은 트랜잭션 입력을 정의. 입력은 이전 트랜잭션에서 돈을 가져오는 것을 나타냄.
type TxIn struct {
	TxID  string `json:"txId"`  // 이전 트랜잭션의 ID. 돈의 출처를 가리킴.
	Index int    `json:"index"` // 이전 트랜잭션의 출력 중 어느 것을 사용하는지 나타냄.
	Owner string `json:"owner"` // 돈을 소유한 사람 (송신자)의 주소.
}

// TxOut은 트랜잭션 출력. 돈이 어디로 가는지 나타냄.
type TxOut struct {
	Owner  string `json:"owner"`  // 돈을 받을 사람의 주소.
	Amount int    `json:"amount"` // 송금할 금액.
}

// UTxOut은 Unspent Transaction Output의 약자. "아직 사용되지 않은 출력"을 나타냄.
type UTxOut struct {
	TxID   string `json:"txId"`   // 트랜잭션 ID. 돈이 어디에서 왔는지 표시.
	Index  int    `json:"index"`  // 출력 목록에서 해당 돈의 위치.
	Amount int    `json:"amount"` // 사용 가능한 금액.
}

// Mempool에 있는 트랜잭션이 특정 UTxO를 사용 중인지 확인하는 함수
func isOnMempool(uTxOut *UTxOut) bool {
	exists := false // 기본값은 사용 중 아님
Outer: // 다중 루프를 빠져나오기 위해 레이블 지정
	for _, tx := range Mempool.Txs { // Mempool의 모든 트랜잭션 순회
		for _, input := range tx.TxIns { // 각 트랜잭션의 입력 목록 확인
			if input.TxID == uTxOut.TxID && input.Index == uTxOut.Index {
				// 동일한 TxID와 Index를 찾으면 사용 중임
				exists = true
				break Outer // 더 이상 확인할 필요 없음
			}
		}
	}
	return exists
}

// Coinbase 트랜잭션은 새로운 블록 생성 시 채굴자에게 보상을 지급하는 특별한 트랜잭션
func makeCoinbaseTx(address string) *Tx {
	txIns := []*TxIn{
		{"", -1, "COINBASE"}, // 입력은 항상 "COINBASE"로 설정. 이전 트랜잭션이 없음.
	}
	txOuts := []*TxOut{
		{address, minerReward}, // 출력: 채굴자의 주소와 보상 금액.
	}
	tx := Tx{
		ID:        "",                  // ID는 나중에 생성
		Timestamp: int(time.Now().Unix()), // 트랜잭션 생성 시간
		TxIns:     txIns,
		TxOuts:    txOuts,
	}
	tx.getId() // 트랜잭션 ID 생성
	return &tx
}

// 일반적인 트랜잭션을 생성하는 함수
func makeTx(from, to string, amount int) (*Tx, error) {
	// 송신자의 잔액 확인. 충분하지 않으면 에러 반환.
	if BalanceByAddress(from, Blockchain()) < amount {
		return nil, errors.New("not enoguh 돈") // "돈이 부족하다"라는 에러 메시지
	}
	var txOuts []*TxOut // 출력 목록
	var txIns []*TxIn   // 입력 목록
	total := 0          // 입력에서 모은 금액의 총합

	// 송신자의 UTxO 목록을 가져와 필요한 만큼 입력을 모음
	uTxOuts := UTxOutsByAddress(from, Blockchain())
	for _, uTxOut := range uTxOuts {
		if total >= amount { // 필요한 금액을 다 모았으면 중단
			break
		}
		txIn := &TxIn{uTxOut.TxID, uTxOut.Index, from} // 입력 생성
		txIns = append(txIns, txIn)
		total += uTxOut.Amount // 입력 총합에 추가
	}

	// 잔돈이 발생하면 송신자에게 다시 돌려주는 출력 생성
	if change := total - amount; change != 0 {
		changeTxOut := &TxOut{from, change} // 잔돈은 송신자로 보냄
		txOuts = append(txOuts, changeTxOut)
	}

	// 수신자에게 보내는 출력 생성
	txOut := &TxOut{to, amount}
	txOuts = append(txOuts, txOut)

	// 최종적으로 트랜잭션을 생성
	tx := &Tx{
		ID:        "",                  // ID는 나중에 생성
		Timestamp: int(time.Now().Unix()), // 트랜잭션 생성 시간
		TxIns:     txIns,
		TxOuts:    txOuts,
	}
	tx.getId() // 트랜잭션 ID 생성
	return tx, nil
}

// Mempool에 새 트랜잭션을 추가하는 함수
func (m *mempool) AddTx(to string, amount int) error {
	// "nico"라는 송신자를 기본값으로 사용
	tx, err := makeTx("nico", to, amount)
	if err != nil {
		return err // 트랜잭션 생성 실패 시 에러 반환
	}
	m.Txs = append(m.Txs, tx) // 트랜잭션을 Mempool에 추가
	return nil
}

// Mempool에 있는 트랜잭션을 블록에 추가할 준비를 하는 함수
func (m *mempool) TxToConfirm() []*Tx {
	coinbase := makeCoinbaseTx("nico") // Coinbase 트랜잭션 생성 (채굴 보상)
	txs := m.Txs                      // Mempool의 모든 트랜잭션 가져오기
	txs = append(txs, coinbase)       // Coinbase 트랜잭션을 추가
	m.Txs = nil                       // Mempool을 비움
	return txs                        // 블록에 포함할 트랜잭션 반환
}
