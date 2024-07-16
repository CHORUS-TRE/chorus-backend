package helper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPass(t *testing.T) {
	var s = "toto"
	h, err := HashPass(s)
	if err != nil {
		t.Fatal(err)
	}

	if !CheckPassHash("$2a$10$AYU7cZCP4ubxSRX72cj6vePp55Ha0MZx6CRDMcny9fjmNlxVSLOmS", s) {
		t.Fail()
	}
	if !CheckPassHash("$2a$10$20w6U8MsKUKR6kypfXHCXe09QJ6R0FDY4hTNlKZwVuaxlxlwcDJYu", s) {
		t.Fail()
	}
	if !CheckPassHash(h, s) {
		t.Fail()
	}
}

func TestGeneratePassword(t *testing.T) {
	str, err := GeneratePassword(20)
	assert.Nil(t, err)
	fmt.Printf("pass: %v\n", str)
	assert.Equal(t, 20, len(str))
}
