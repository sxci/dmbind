package pub

import (
	"testing"
	"strconv"
)

func TestPubAll(t *testing.T) {
	// Setup(&Mac {
		// AccessKey: "tGf47MBl1LyT9uaNv-NZV4XZe7sKxOIa9RE2Lp8B",
		// SecretKey: []byte("zhbiA6gcQMEi22uZ8CBGvmbnD2sR8SO-5S8qlLCG"),
	// })

	// t.Error(Publish("us-archive-ubuntu.qiniudn.com", "us-archive-ubuntu"))
	return
	for i:=101; i<128; i++ {
		Publish("ali-test" + strconv.Itoa(i) + ".qiniudn.com", "demo")
	}
}
