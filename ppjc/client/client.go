package ppjc

import (
	"context"
	"errors"
	"net/http"

	"net/url"

	"strings"

	"strconv"

	"io/ioutil"

	"encoding/json"

	"io"

	"sync"

	"mime/multipart"

	"github.com/Sirupsen/logrus"
	"github.com/cs3238-tsuzu/chan-utils"
	"github.com/cs3238-tsuzu/popcon-sc/lib/database"
	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/cs3238-tsuzu/popcon-sc/ppjc/types"
	"github.com/gorilla/websocket"
)

type Client struct {
	addr, auth string
}

func (client *Client) ContestsRanking(cid, limit, offset int64) ([]database.RankingRow, error) {
	u, err := url.Parse(client.addr)

	if err != nil {
		return nil, err
	}
	u.Path = "/v1/contests/" + strconv.FormatInt(cid, 10) + "/ranking"

	val := url.Values{}
	val.Add("limit", strconv.FormatInt(limit, 10))
	val.Add("offset", strconv.FormatInt(offset, 10))

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(val.Encode()))

	if err != nil {
		return nil, err
	}

	req.Header.Set(sctypes.InternalHTTPToken, client.auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("error: " + resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var rows []database.RankingRow
	if err := json.Unmarshal(b, &rows); err != nil {
		return nil, err
	}

	return rows, nil
}

func (client *Client) ContestsNew(cid int64) error {
	u, err := url.Parse(client.addr)

	if err != nil {
		return err
	}
	u.Path = "/v1/contests/" + strconv.FormatInt(cid, 10) + "/new"

	req, err := http.NewRequest("POST", u.String(), nil)

	if err != nil {
		return err
	}

	req.Header.Set(sctypes.InternalHTTPToken, client.auth)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("error: " + resp.Status)
	}

	return nil
}

func (client *Client) ContestsProblemsAdd(cid, pid int64) error {
	u, err := url.Parse(client.addr)

	if err != nil {
		return err
	}
	u.Path = "/v1/contests/" + strconv.FormatInt(cid, 10) + "/problems/add"

	val := url.Values{}
	val.Add("pid", strconv.FormatInt(pid, 10))

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(val.Encode()))

	if err != nil {
		return err
	}

	req.Header.Set(sctypes.InternalHTTPToken, client.auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("error: " + resp.Status)
	}
	return nil
}

func (client *Client) ContestsProblemsDelete(cid, pid int64) error {
	u, err := url.Parse(client.addr)

	if err != nil {
		return err
	}
	u.Path = "/v1/contests/" + strconv.FormatInt(cid, 10) + "/problems/delete"

	val := url.Values{}
	val.Add("pid", strconv.FormatInt(pid, 10))

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(val.Encode()))

	if err != nil {
		return err
	}

	req.Header.Set(sctypes.InternalHTTPToken, client.auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("error: " + resp.Status)
	}
	return nil
}

func (client *Client) ContestsJoin(cid, iid int64) error {
	u, err := url.Parse(client.addr)

	if err != nil {
		return err
	}
	u.Path = "/v1/contests/" + strconv.FormatInt(cid, 10) + "/join"

	val := url.Values{}
	val.Add("iid", strconv.FormatInt(iid, 10))

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(val.Encode()))

	if err != nil {
		return err
	}

	req.Header.Set(sctypes.InternalHTTPToken, client.auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("error: " + resp.Status)
	}
	return nil
}

func (client *Client) JudgeSubmit(cid, sid int64) error {
	u, err := url.Parse(client.addr)

	if err != nil {
		return err
	}
	u.Path = "/v1/judge/submit"

	val := url.Values{}
	val.Add("cid", strconv.FormatInt(cid, 10))
	val.Add("sid", strconv.FormatInt(sid, 10))

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(val.Encode()))

	if err != nil {
		return err
	}

	req.Header.Set(sctypes.InternalHTTPToken, client.auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("error: " + resp.Status)
	}
	return nil
}

func (client *Client) JudgeSubmissionsUpdateCase(cid, sid, jid int64, status string, res database.SubmissionTestCase) error {
	u, err := url.Parse(client.addr)

	if err != nil {
		return err
	}
	u.Path = "/v1/judge/submissions/updateCase"

	b, _ := json.Marshal(ppjctypes.JudgeTestcaseResult{
		Cid:      cid,
		Sid:      sid,
		Jid:      jid,
		Testcase: res,
	})

	val := url.Values{}
	val.Add("testcase_result", string(b))

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(val.Encode()))

	if err != nil {
		return err
	}

	req.Header.Set(sctypes.InternalHTTPToken, client.auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("error: " + resp.Status)
	}
	return nil
}

func (client *Client) JudgeSubmissionsUpdateResult(cid, sid, jid int64, status sctypes.SubmissionStatusType, score int64, time, mem int64, message io.Reader) error {
	u, err := url.Parse(client.addr)

	if err != nil {
		return err
	}
	u.Path = "/v1/judge/submissions/updateResult"

	b, _ := json.Marshal(ppjctypes.JudgeSubmissionResult{
		Cid:    cid,
		Sid:    sid,
		Jid:    jid,
		Status: status,
		Score:  score,
		Time:   time,
		Mem:    mem,
	})

	pread, pwrite := io.Pipe()
	writer := multipart.NewWriter(pwrite)

	go func() {
		writer.WriteField("submission_result", string(b))

		if message != nil {
			msg, _ := writer.CreateFormField("message")
			io.Copy(msg, message)
		}
		writer.Close()
	}()
	req, err := http.NewRequest("POST", u.String(), pread)

	if err != nil {
		return err
	}
	req.Header.Set(sctypes.InternalHTTPToken, client.auth)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("error: " + resp.Status)
	}
	return nil
}

func (client *Client) FileDownload(category, name string) (io.ReadCloser, error) {
	u, err := url.Parse(client.addr)

	if err != nil {
		return nil, err
	}
	u.Path = "/v1/file_download"

	val := url.Values{}
	val.Add("category", category)
	val.Add("name", name)

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(val.Encode()))

	if err != nil {
		return nil, err
	}

	req.Header.Set(sctypes.InternalHTTPToken, client.auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, errors.New("error: " + resp.Status)
	}
	return resp.Body, nil
}

func (client *Client) StartWorkersWSPolling(parallel int64, judgeInfoChan chan<- ppjctypes.JudgeInformation, ctx context.Context) (<-chan error, func(), error) {
	u, err := url.Parse(client.addr)

	if err != nil {
		return nil, nil, err
	}
	u.Path = "/v1/workers/ws/polling"
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{"Popcon-Parallel-Judge": []string{strconv.FormatInt(parallel, 10)}})

	if err != nil {
		return nil, nil, err
	}

	exitChan := make(chan error, 1)
	exited := chanUtils.NewExitedNotifier()
	wg := sync.WaitGroup{}
	wg.Add(1)

	messageSendingTrigger := chanUtils.NewTrigger()
	go func() {
		defer wg.Done()
		for {
			ctx, canceller := context.WithCancel(context.Background())

			done := make(chan bool)
			go func() {
				defer close(done)
				if err := messageSendingTrigger.WaitWithContext(ctx); err == nil {
					if err := conn.WriteJSON(ppjctypes.JudgeOneFinished); err != nil {
						logrus.WithError(err).Error("WriteJSON error()")
						return
					}
				}
			}()
			select {
			case <-exited.Channel:
				canceller()
				return
			case <-done:
			}
			canceller()
		}
	}()

	ret := func() {
		messageSendingTrigger.Wake()
	}

	go func() {
		defer close(judgeInfoChan)
		var exitError error
		for {
			var info ppjctypes.JudgeInformation
			err := conn.ReadJSON(&info)

			if err != nil {
				exitError = err
				exited.Finish()
				break
			}

			judgeInfoChan <- info

			select {
			case <-ctx.Done():
				exited.Finish()
				exitError = ctx.Err()
			default:
			}
		}
		wg.Wait()
		exitChan <- exitError
	}()

	return exitChan, ret, nil
}

func NewClient(addr, auth string) (*Client, error) {
	_, err := url.Parse(addr)

	if err != nil {
		return nil, err
	}

	return &Client{
		addr: addr,
		auth: auth,
	}, nil
}
