//go:build unit || integration || acceptance

package main

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
	_ "github.com/lib/pq"
)

// need to set env TEST_CONFIG_FILE with "./configs/dev/chorus.yaml" if run from base
// i.e. TEST_CONFIG_FILE="./configs/dev/chorus.yaml" go run --tags=unit ./tests/steward/getadmintoken/main.go
func main() {
	helpers.Setup()

	token := helpers.CreateJWTToken(1, 88888, authorization.RoleSuperAdmin.String(), map[string]string{"user": "*", "workspace": "*", "workbench": "*"})

	fmt.Println("token", token)
}
