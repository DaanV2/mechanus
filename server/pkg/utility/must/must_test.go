package must_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/pkg/utility/must"
)

var _ = Describe("Must", func() {
	Context("Do", func() {
		It("should return the value when no error", func() {
			result := must.Do(42, nil)
			Expect(result).To(Equal(42))

			str := must.Do("hello", nil)
			Expect(str).To(Equal("hello"))
		})

		It("should panic when error is not nil", func() {
			testError := errors.New("test error")

			Expect(func() {
				must.Do(42, testError)
			}).To(PanicWith(testError))
		})

		It("should work with different types", func() {
			type testStruct struct {
				Value string
			}

			test := testStruct{Value: "test"}
			result := must.Do(test, nil)
			Expect(result).To(Equal(test))
		})

		It("should work with time parsing", func() {
			// Test successful parsing
			expectedTime := must.Do(time.Parse("15:04", "12:00"))
			Expect(expectedTime.Format("15:04")).To(Equal("12:00"))

			// Test panic on invalid time
			Expect(func() {
				must.Do(time.Parse("15:04", "invalid"))
			}).To(PanicWith(MatchError(ContainSubstring("cannot parse"))))
		})
	})
})
