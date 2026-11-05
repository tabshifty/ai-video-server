package cloud

import "testing"

func TestIsPanPageMatchesDirectAndGotoURLs(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"https://115.com/?cid=0&offset=0&mode=wangpan", true},
		{"https://115.com/?goto=http%3A%2F%2F115.com%2F%3Fcid%3D0%26offset%3D0%26mode%3Dwangpan", true},
		{"https://115.com/?ct=file", true},
		{"https://115.com/storage/allfiles", true},
		{"https://www.115.com/?mode=wangpan", true},
		{"https://115.com/", false},
		{"https://115.com/login", false},
		{"https://yun.115.com/chat", false},
		{"https://example.com/?mode=wangpan", false},
		{"chrome://transfer-frame/", false},
	}
	for _, tc := range cases {
		if got := isPanPage(Target{Type: "page", URL: tc.url}); got != tc.want {
			t.Errorf("isPanPage(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
	if isPanPage(Target{Type: "shared_worker", URL: "https://115.com/?mode=wangpan"}) {
		t.Fatal("non-page targets must not be treated as the pan page")
	}
}

func TestIs115PageStaysOnAppHosts(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"https://115.com/", true},
		{"https://www.115.com/?mode=wangpan", true},
		{"https://yun.115.com/chat", false},
		{"https://cdnassets.115.com/login/loginworker/loginWkApi.js", false},
		{"https://115.com.evil.example/", false},
	}
	for _, tc := range cases {
		if got := is115Page(Target{Type: "page", URL: tc.url}); got != tc.want {
			t.Errorf("is115Page(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
}
