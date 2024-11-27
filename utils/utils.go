package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)

// HandleErr는 에러가 nil이 아닐 경우 로그를 남기고 패닉을 발생시킵니다.
func HandleErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// ToBytes는 인터페이스를 gob 인코딩을 사용하여 바이트 슬라이스로 변환합니다.
func ToBytes(i interface{}) []byte {
	var aBuffer bytes.Buffer
	encoder := gob.NewEncoder(&aBuffer)
	HandleErr(encoder.Encode(i)) // 인코딩 에러를 처리합니다.
	return aBuffer.Bytes()
}

// FromBytes는 바이트 슬라이스를 gob 디코딩을 사용하여 인터페이스로 변환합니다.
func FromBytes(i interface{}, data []byte) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	HandleErr(decoder.Decode(i)) // 디코딩 에러를 처리합니다.
}

// Hash는 인터페이스의 문자열 표현에 대한 SHA-256 해시를 반환합니다.
func Hash(i interface{}) string {
	s := fmt.Sprintf("%v", i) // 인터페이스를 문자열로 변환합니다.
	hash := sha256.Sum256([]byte(s)) // SHA-256 해시를 계산합니다.
	return fmt.Sprintf("%x", hash) // 해시를 16진수 문자열로 반환합니다.
}