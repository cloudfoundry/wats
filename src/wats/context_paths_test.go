package wats

import (
	"fmt"
	"strings"

	. "code.cloudfoundry.org/cf-routing-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Context Paths", func() {
	var (
		appName1 string

		appName2 string
		app2Path = "/app2"
		appName3 string
		app3Path = "/app3/long/sub/path"
		hostname string
	)

	BeforeEach(func() {
		if !config.ContextPath {
			Skip("Skipping ContextPath tests")
		}
		domain := config.GetAppsDomain()

		appName1 = generator.PrefixedRandomName(config.GetNamePrefix(), "APP")
		Expect(pushNora(appName1).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
		Expect(cf.Cf("start", appName1).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

		appName2 = generator.PrefixedRandomName(config.GetNamePrefix(), "APP")
		Expect(pushNoraWithNoRoute(appName2).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

		appName3 = generator.PrefixedRandomName(config.GetNamePrefix(), "APP")
		Expect(pushNoraWithNoRoute(appName3).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))

		hostname = appName1

		MapRouteToApp(appName2, domain, hostname, app2Path, CF_PUSH_TIMEOUT)
		MapRouteToApp(appName3, domain, hostname, app3Path, CF_PUSH_TIMEOUT)

		Expect(cf.Cf("start", appName2).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
		Expect(cf.Cf("start", appName3).Wait(CF_PUSH_TIMEOUT)).To(gexec.Exit(0))
	})

	AfterEach(func() {
		AppReport(appName1, CF_PUSH_TIMEOUT)
		AppReport(appName2, CF_PUSH_TIMEOUT)
		AppReport(appName3, CF_PUSH_TIMEOUT)

		DeleteApp(appName1, CF_PUSH_TIMEOUT)
		DeleteApp(appName2, CF_PUSH_TIMEOUT)
		DeleteApp(appName3, CF_PUSH_TIMEOUT)
	})

	Context("when another app has a route with a context path", func() {
		It("routes to app with context path", func() {
			Eventually(func() string {
				return helpers.CurlAppRoot(config, hostname)
			}, CF_PUSH_TIMEOUT).Should(ContainSubstring(strings.ToLower(appName1)))

			Eventually(func() string {
				return helpers.CurlApp(config, hostname, fmt.Sprintf("%s/env/VCAP_APPLICATION", app2Path))
			}, CF_PUSH_TIMEOUT).Should(ContainSubstring(fmt.Sprintf(`\"application_name\":\"%s\"`, appName2)))

			Eventually(func() string {
				return helpers.CurlApp(config, hostname, fmt.Sprintf("%s/env/VCAP_APPLICATION", app3Path))
			}, CF_PUSH_TIMEOUT).Should(ContainSubstring(fmt.Sprintf(`\"application_name\":\"%s\"`, appName3)))
		})
	})
})
