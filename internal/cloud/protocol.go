package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ErrCodeTaskExists is 115's "task already exists" response. It is a success
// for our purposes: the same link is already queued or finished.
const ErrCodeTaskExists = 10008

// PanPageURL is the page that carries the native bridge and USER_ID.
const PanPageURL = "https://115.com/?cid=0&offset=0&mode=wangpan"

// State is the observable browser session state.
type State struct {
	Up       bool
	LoggedIn bool
	UserID   string
	PageURL  string
	Quota    *Quota
}

// Quota is the cloud-download monthly quota when 115 reports it.
type Quota struct {
	Total   int64 `json:"total"`
	Used    int64 `json:"used"`
	Surplus int64 `json:"surplus"`
	// Known reports that 115 returned at least one quota field.
	Known bool `json:"known"`
	// SurplusKnown distinguishes a real zero surplus from a missing field;
	// only a known exhausted quota may block submission.
	SurplusKnown bool `json:"surplus_known"`
}

// Task is one cloud-download task as reported by 115.
type Task struct {
	InfoHash   string  `json:"info_hash"`
	Name       string  `json:"name"`
	Size       int64   `json:"size"`
	Status     string  `json:"status"`
	StatusText string  `json:"status_text"`
	Percent    float64 `json:"percent"`
	PickCode   string  `json:"pick_code"`
	FileID     string  `json:"file_id"`
	AddTime    int64   `json:"add_time"`
}

// Finished reports whether 115 says the task completed.
func (t Task) Finished() bool {
	return t.Status == "finished" || t.Percent >= 100
}

// SubmitResult is the add_task_urls response.
type SubmitResult struct {
	State    bool   `json:"state"`
	ErrCode  int    `json:"errcode"`
	ErrorMsg string `json:"error_msg"`
}

// Probe reads the page-level capabilities and login state.
func Probe(ctx context.Context, s *Session) (State, error) {
	const expr = `(() => ({
		url: location.href,
		lixianKey: typeof (window.browserInterface && browserInterface.ChromeGetLixianKey) === 'function',
		download: typeof (window.browserInterface && browserInterface.CreateDownloadTask) === 'function',
		userId: (typeof USER_ID !== 'undefined' && USER_ID) ? String(USER_ID) : null
	}))()`
	var raw struct {
		URL       string `json:"url"`
		LixianKey bool   `json:"lixianKey"`
		Download  bool   `json:"download"`
		UserID    string `json:"userId"`
	}
	if err := s.Evaluate(ctx, expr, &raw); err != nil {
		return State{}, err
	}
	state := State{
		Up:       true,
		LoggedIn: raw.UserID != "",
		UserID:   raw.UserID,
		PageURL:  raw.URL,
	}
	if !raw.LixianKey || !raw.Download {
		return state, fmt.Errorf("cloud: page %s is missing the 115 native bridge", raw.URL)
	}
	return state, nil
}

// QuotaInfo reads the monthly cloud-download quota. Unknown shapes are not an
// error: the caller only uses a known-exhausted quota to block submission.
func QuotaInfo(ctx context.Context, s *Session) (Quota, error) {
	const expr = `(async () => {
		const pick = (...vals) => { for (const v of vals) if (typeof v === 'number') return v; return null; };
		try {
			const r = await fetch('/web/lixian/?ct=lixian&ac=get_quota_package_info', { credentials: 'include' });
			const j = await r.json();
			const d = j.data || {};
			return {
				total: pick(j.total, j.count, j.total_count, d.total, d.count),
				used: pick(j.used, j.used_count, d.used, d.used_count),
				surplus: pick(j.surplus, j.remain, j.remaining, d.surplus, d.remain, d.remaining)
			};
		} catch (e) {
			return { total: null, used: null, surplus: null };
		}
	})()`
	var raw struct {
		Total   *int64 `json:"total"`
		Used    *int64 `json:"used"`
		Surplus *int64 `json:"surplus"`
	}
	if err := s.Evaluate(ctx, expr, &raw); err != nil {
		return Quota{}, err
	}
	quota := Quota{}
	if raw.Total != nil {
		quota.Total = *raw.Total
		quota.Known = true
	}
	if raw.Used != nil {
		quota.Used = *raw.Used
		quota.Known = true
	}
	if raw.Surplus != nil {
		quota.Surplus = *raw.Surplus
		quota.SurplusKnown = true
		quota.Known = true
	}
	return quota, nil
}

// TargetDirID resolves the cloud-download target directory id.
func TargetDirID(ctx context.Context, s *Session) (string, error) {
	const expr = `(async () => {
		const r = await fetch('/?ct=lixian&ac=get_id', { credentials: 'include' });
		const j = await r.json();
		return { state: !!j.state, cid: j.cid ? String(j.cid) : '' };
	})()`
	var raw struct {
		State bool   `json:"state"`
		CID   string `json:"cid"`
	}
	if err := s.Evaluate(ctx, expr, &raw); err != nil {
		return "", err
	}
	if raw.CID == "" {
		return "", fmt.Errorf("cloud: cannot resolve cloud download directory (state=%v)", raw.State)
	}
	return raw.CID, nil
}

// SubmitURLs submits links to 115 cloud download in one batch.
func SubmitURLs(ctx context.Context, s *Session, links []string, wpPathID string) (SubmitResult, error) {
	linksJSON, err := json.Marshal(links)
	if err != nil {
		return SubmitResult{}, err
	}
	expr := fmt.Sprintf(`(async () => {
		const body = new URLSearchParams();
		body.set('savepath', '');
		body.set('wp_path_id', %s);
		const links = %s;
		links.forEach((l, i) => body.append('url[' + i + ']', l));
		const r = await fetch('/web/lixian/?ct=lixian&ac=add_task_urls', {
			method: 'POST', credentials: 'include',
			headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
			body: body.toString()
		});
		const text = await r.text();
		let j;
		try { j = JSON.parse(text); } catch (e) {
			return { state: false, errcode: -1, error_msg: 'non-json: ' + text.slice(0, 180) };
		}
		return { state: !!j.state, errcode: j.errcode || 0, error_msg: j.error_msg || j.msg || '' };
	})()`, jsString(wpPathID), string(linksJSON))

	var out SubmitResult
	if err := s.Evaluate(ctx, expr, &out); err != nil {
		return SubmitResult{}, err
	}
	return out, nil
}

// ListTasks returns one page of 115 cloud-download tasks for a stat filter.
func ListTasks(ctx context.Context, s *Session, stat int, page int) ([]Task, error) {
	expr := fmt.Sprintf(`(async () => {
		const body = new URLSearchParams();
		body.set('page', %s);
		body.set('stat', %s);
		try {
			const t = String(Math.floor(Date.now() / 1000));
			const sign = browserInterface.ChromeGetLixianKey(t, String(USER_ID));
			body.set('uid', String(USER_ID));
			body.set('sign', sign);
			body.set('time', t);
		} catch (e) {}
		const r = await fetch('/web/lixian/?ct=lixian&ac=task_lists', {
			method: 'POST', credentials: 'include',
			headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
			body: body.toString()
		});
		const j = await r.json();
		return (j.tasks || []).map(x => ({
			info_hash: String(x.info_hash || '').toLowerCase(),
			name: String(x.name || ''),
			size: Number(x.size) || 0,
			status: String(x.display_status || ''),
			status_text: String(x.status_text || ''),
			percent: Number(x.display_percent) || 0,
			pick_code: String(x.pick_code || ''),
			file_id: x.file_id ? String(x.file_id) : '',
			add_time: Number(x.add_time) || 0
		}));
	})()`, jsString(fmt.Sprint(page)), jsString(fmt.Sprint(stat)))

	var tasks []Task
	if err := s.Evaluate(ctx, expr, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// FindTask locates a task by info hash across the downloading and finished
// lists. Page 1 is enough: 115 orders tasks newest first and a just-submitted
// task is at the top.
func FindTask(ctx context.Context, s *Session, infoHash string) (*Task, error) {
	want := normalizeHash(infoHash)
	for _, stat := range []int{12, 13} {
		tasks, err := ListTasks(ctx, s, stat, 1)
		if err != nil {
			return nil, err
		}
		for i := range tasks {
			if normalizeHash(tasks[i].InfoHash) == want {
				return &tasks[i], nil
			}
		}
	}
	return nil, nil
}

// CreateDownloadTask asks the native downloader to pull a finished cloud file.
func CreateDownloadTask(ctx context.Context, s *Session, name, pickCode string) error {
	expr := fmt.Sprintf(`(() => {
		if (typeof browserInterface.CreateDownloadTask !== 'function') {
			throw new Error('CreateDownloadTask is unavailable');
		}
		const payload = { list: [{ is_dir: false, n: %s, pc: %s }], count: 1, ref_url: location.href };
		browserInterface.CreateDownloadTask(encodeURIComponent(JSON.stringify(payload)));
		return true;
	})()`, jsString(name), jsString(pickCode))
	var ok bool
	if err := s.Evaluate(ctx, expr, &ok); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("cloud: CreateDownloadTask did not run")
	}
	return nil
}

// WaitForTask polls 115 until the task reaches a terminal or usable state.
// It returns the last observed task so callers can report progress.
func WaitForTask(ctx context.Context, s *Session, infoHash string, interval time.Duration, onProgress func(Task)) (*Task, error) {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	var last *Task
	for {
		task, err := FindTask(ctx, s, infoHash)
		if err != nil {
			return last, err
		}
		if task != nil {
			last = task
			if onProgress != nil {
				onProgress(*task)
			}
			if task.Finished() {
				return task, nil
			}
			if task.Status == "failed" || task.Status == "deleted" {
				return task, fmt.Errorf("cloud: 115 task %s (%s)", task.Status, task.StatusText)
			}
		}
		select {
		case <-ctx.Done():
			if last != nil {
				return last, ctx.Err()
			}
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
}

func normalizeHash(value string) string {
	out := make([]rune, 0, len(value))
	for _, r := range value {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		out = append(out, r)
	}
	return string(out)
}

// jsString quotes a value as a JavaScript string literal.
func jsString(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return `""`
	}
	return string(encoded)
}
