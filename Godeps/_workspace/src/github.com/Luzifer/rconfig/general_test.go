package rconfig

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing general parsing", func() {
	type t struct {
		Test        string `default:"foo" env:"shell" flag:"shell" description:"Test"`
		Test2       string `default:"blub" env:"testvar" flag:"testvar,t" description:"Test"`
		DefaultFlag string `default:"goo"`
		SadFlag     string
	}

	var (
		err  error
		args []string
		cfg  t
	)

	Context("with defined arguments", func() {
		BeforeEach(func() {
			cfg = t{}
			args = []string{
				"--shell=test23",
				"-t", "bla",
			}
		})

		JustBeforeEach(func() {
			err = parse(&cfg, args)
		})

		It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
		It("should have parsed the expected values", func() {
			Expect(cfg.Test).To(Equal("test23"))
			Expect(cfg.Test2).To(Equal("bla"))
			Expect(cfg.SadFlag).To(Equal(""))
			Expect(cfg.DefaultFlag).To(Equal("goo"))
		})
	})

	Context("with no arguments", func() {
		BeforeEach(func() {
			cfg = t{}
			args = []string{}
		})

		JustBeforeEach(func() {
			err = parse(&cfg, args)
		})

		It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
		It("should have used the default value", func() {
			Expect(cfg.Test).To(Equal("foo"))
		})
	})

	Context("with no arguments and set env", func() {
		BeforeEach(func() {
			cfg = t{}
			args = []string{}
			os.Setenv("shell", "test546")
		})

		AfterEach(func() {
			os.Unsetenv("shell")
		})

		JustBeforeEach(func() {
			err = parse(&cfg, args)
		})

		It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
		It("should have used the value from env", func() {
			Expect(cfg.Test).To(Equal("test546"))
		})
	})

})
