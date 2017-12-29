package debug

import (
	"fmt"
	"net/http"
)

var (
	// APP 当前应用名
	APP = ""
	// GitTime git log中记录的提交时间.
	GitTime = ""
	// GitHash git commit hash.
	GitHash = ""
	// GitMessage git log 中记录的提交信息.
	GitMessage = ""
)

// Print 输出当前程序编译信息.
func Print() {
	fmt.Printf("DBS - %s\n", APP)
	fmt.Printf("Commit Hash: %s\n", GitHash)
	fmt.Printf("Commit Time: %s\n", GitTime)
	fmt.Printf("Commit Message: %s\n", GitMessage)
}

type Version struct {
}

func (v *Version) GET(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "DBS - %s\n", APP)
	fmt.Fprintf(w, "Commit Hash: %s\n", GitHash)
	fmt.Fprintf(w, "Commit Time: %s\n", GitTime)
	fmt.Fprintf(w, "Commit Message: %s\n", GitMessage)
}
