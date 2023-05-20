package cmd

import (
	"fmt"
	"testing"
)

func TestPrintApplyResult(t *testing.T) {

	printApplyResult(Apply_Kind_Pod, ApplyResult_Success, "created", "apiserver code:201")
	fmt.Println()
	printApplyResult(Apply_Kind_Pod, ApplyResult_Failed, "failed", "apiserver code:404")
	fmt.Println()
	printApplyResult(Apply_Kind_Pod, ApplyResult_Unknow, "created", "apiserver code:201")
	fmt.Println()
}

func TestPrintApplyObjInfo(t *testing.T) {
	printApplyObjectInfo(Apply_Kind_Pod, "pod1", "default")
	fmt.Println()
	printApplyObjectInfo(Apply_Kind_Pod, "pod2", "default")
	fmt.Println()
	printApplyObjectInfo(Apply_Kind_Pod, "pod3", "default")
}
