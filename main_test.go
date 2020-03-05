package main

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
)

func Test_loadConfig(t *testing.T) {
	loadConfig()

	assert.Equal(t, true, viper.GetBool("delete-empty-dir"), "they should be equal")
	assert.Equal(t, 100*time.Millisecond, viper.GetDuration("interval"))

	viper.Set("testSlice", []string{"a", "b"})
	assert.EqualValues(t, []string{"a", "b"}, viper.GetStringSlice("testSlice"), "should be equal")
}

func Test_cmdline(t *testing.T) {
	var pkgdir = flag.String("pkgdir", "/pkgdirs", "dir of package containing embedded files")

	t.Log(*pkgdir)

	os.Args = append(os.Args, `/vods`)
	loadConfig()
	t.Log(viper.GetStringSlice("dirs"))
}
