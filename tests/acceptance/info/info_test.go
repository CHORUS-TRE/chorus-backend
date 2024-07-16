//go:build acceptance

package info_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("chorus infos", func() {
	helpers.Setup()

	Describe("get the info of chorus", func() {

		Given("a browser/curl", func() {

			When("the route 'GET /' is called", func() {
				defer GinkgoRecover()

				infosReply, err := getInfos()

				Then("it should not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				Then("it should return the info of the chorus", func() {
					Expect(infosReply.Name).Should(Equal("chorus"))
					Expect(infosReply.GoVersion).Should(BeZero())
				})
			})
		})
	})
})

func getInfos() (*provider.Info, error) {
	req, err := http.NewRequest("GET", "http://"+helpers.ComponentURL(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("wrong status code: %v", resp.StatusCode)
	}

	var infosReply provider.Info
	err = json.NewDecoder(resp.Body).Decode(&infosReply)
	if err != nil {
		return nil, err
	}

	return &infosReply, nil
}
