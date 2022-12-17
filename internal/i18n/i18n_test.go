package i18n

import (
	"fmt"
	"testing"

	"github.com/hatlonely/go-kit/strx"
	. "github.com/smartystreets/goconvey/convey"
)

func TestI18n(t *testing.T) {
	Convey("TestI18n", t, func() {
		i18n := &I18n{
			Title: Title{
				Report:  "test-report",
				Summary: "test-summary",
			},
			Tooltip: Tooltip{
				Save: "test-save",
			},
		}

		merge(i18n, defaultI18n["dft"])

		fmt.Println(strx.JsonMarshalIndentSortKeys(i18n))
	})
}
