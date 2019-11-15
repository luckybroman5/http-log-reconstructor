package statsd

import (
	"testing"

	"github.com/luckybroman5/http-log-reconstructor/k6/stats"
	"github.com/luckybroman5/http-log-reconstructor/k6/stats/statsd/common/testutil"
	"github.com/stretchr/testify/require"
)

func TestCollector(t *testing.T) {
	testutil.BaseTest(t, New,
		func(t *testing.T, _ []stats.SampleContainer, expectedOutput, output string) {
			require.Equal(t, expectedOutput, output)
		})
}
