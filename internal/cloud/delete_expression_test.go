package cloud

import (
	"strings"
	"testing"
)

func TestDeleteFileExpressionUsesActualFileIdentity(t *testing.T) {
	const parentID = "2726135308640032912"
	const fileID = "3544176998614042374"
	expr := deleteFileExpression(parentID, fileID)

	for _, want := range []string{
		`https://webapi.115.com/rb/delete`,
		`body.set('pid', "` + parentID + `")`,
		`body.set('fid[0]', "` + fileID + `")`,
		`body.set('ignore_warn', '1')`,
		`j.state === true || j.state === 1 || j.state === '1'`,
		`json_parsed`,
	} {
		if !strings.Contains(expr, want) {
			t.Fatalf("delete expression missing %q", want)
		}
	}
	if strings.Count(expr, "https://webapi.115.com/rb/delete") != 1 {
		t.Fatalf("delete endpoint must appear exactly once")
	}
	if strings.Contains(expr, "fid[0]', \""+parentID+"\"") {
		t.Fatalf("parent id must not be sent as fid[0]")
	}
}
