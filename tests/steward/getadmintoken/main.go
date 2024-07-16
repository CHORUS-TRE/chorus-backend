//go:build unit || integration || acceptance

package main

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
	_ "github.com/lib/pq"
)

// need to set env TEST_CONFIG_FILE with "./configs/dev/chorus.yml" if run from base
// i.e. TEST_CONFIG_FILE="./configs/dev/chorus.yml" go run --tags=unit ./tests/steward/getadmintoken/main.go
func main() {
	helpers.Setup()

	token := helpers.CreateJWTToken(1, 88888, model.RoleChorus.String())

	fmt.Println("token", token)
}
