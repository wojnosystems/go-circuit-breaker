package circuitHTTP_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/wojnosystems/go-circuit-breaker/circuitHTTP"
	"github.com/wojnosystems/go-circuit-breaker/twoStateCircuit"
	"net/http"
	"net/url"
)

const badURL = "\a"

var _ = Describe("Client", func() {
	var (
		server *ghttp.Server
		client *circuitHTTP.Client
	)
	BeforeEach(func() {
		server = ghttp.NewServer()
		client = circuitHTTP.New(twoStateCircuit.New(twoStateCircuit.Opts{}), http.DefaultClient)
	})
	AfterEach(func() {
		server.Close()
	})
	When("get", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.RespondWith(http.StatusOK, nil),
			)
		})
		It("succeeds", func() {
			resp, err := client.Get(server.URL())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		})
		It("fails to parse the URL", func() {
			_, err := client.Get(badURL)
			Expect(err).Should(HaveOccurred())
		})
	})
	When("head", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.RespondWith(http.StatusOK, nil),
			)
		})
		It("succeeds", func() {
			resp, err := client.Head(server.URL())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		})
		It("fails to parse the URL", func() {
			_, err := client.Head(badURL)
			Expect(err).Should(HaveOccurred())
		})
	})
	When("post", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.RespondWith(http.StatusOK, nil),
			)
		})
		It("succeeds", func() {
			resp, err := client.Post(server.URL(), "application/json", nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		})
		It("fails to parse the URL", func() {
			_, err := client.Post(badURL, "application/json", nil)
			Expect(err).Should(HaveOccurred())
		})
	})
	When("post form", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.RespondWith(http.StatusOK, nil),
			)
		})
		It("succeeds", func() {
			resp, err := client.PostForm(server.URL(), url.Values{})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		})
	})
})
