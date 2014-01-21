package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "github.com/elazarl/goproxy"
    "io/ioutil"
    "kcsapi"
    "log"
    . "net/http"
    "net/url"
    "os"
    "regexp"
    "strings"
    "time"
)

const (
    LOG_FILENAME = "log.txt"
)

var (
    repairingNdockId = make(map[int16]bool)
)

func notifyMissionComplete(d time.Duration) {
    time.Sleep(d)
    PostForm(IM_KAYAC_URL, url.Values{"message": {"艦隊が遠征から帰還しました"}})
}

func notifyRepairComplete(d time.Duration, NdockId int16) {
    if repairingNdockId[NdockId] != true {
        repairingNdockId[NdockId] = true
        time.Sleep(d)
        PostForm(IM_KAYAC_URL, url.Values{"message": {"艦娘の修理が完了しました"}})
        delete(repairingNdockId, NdockId)
    }
}

func assertJSON(data interface{}, file *os.File) {
    switch data.(type) {
    case string:
        file.Write([]byte(data.(string)))
        file.Write([]byte("\n"))
    case float64:
        file.Write([]byte(fmt.Sprintf("%f", data.(float64))))
    default:
    }
}

func main() {
    proxy := goproxy.NewProxyHttpServer()
    proxy.Verbose = true
    f, err := os.Create(LOG_FILENAME)
    if err != nil {
        panic(err)
    }

    re := regexp.MustCompile(`kcsapi`)
    proxy.OnResponse().DoFunc(
        func(resp *Response, ctx *goproxy.ProxyCtx) *Response {
            r := resp.Request
            path := r.URL.Path
            if re.MatchString(path) == true {
                pathData := []byte(path + "\n")
                f.Write(pathData)
                reader := bufio.NewReader(resp.Body)
                b, _ := ioutil.ReadAll(reader)
                buf := bytes.NewBuffer(b)
                resp.Body = ioutil.NopCloser(buf)

                strBody := string(b[:])
                jsonBody := strings.Replace(strBody, "svdata=", "", 1)

                switch path {
                case "/kcsapi/api_get_member/basic":
                    var d kcsapi.BasicData
                    json.Unmarshal([]byte(jsonBody), &d)
                    f.Write([]byte(fmt.Sprintf("%+v\n", d)))
                case "/kcsapi/api_get_member/ndock":
                    var d kcsapi.NdockData
                    json.Unmarshal([]byte(jsonBody), &d)
                    f.Write([]byte(fmt.Sprintf("%+v\n", d)))
                    f.Write([]byte("\n"))
                    for i := 0; i < len(d.ApiData); i++ {
                        timeEpoch := d.ApiData[i].CompleteTime / 1000
                        if timeEpoch != 0 {
                            duration := time.Unix(timeEpoch, 0).Sub(time.Now())
                            go notifyRepairComplete(duration, d.ApiData[i].Id)
                        }
                    }

                case "/kcsapi/api_req_mission/start":
                    var d kcsapi.MissionStartData
                    json.Unmarshal([]byte(jsonBody), &d)
                    f.Write([]byte(fmt.Sprintf("%+v\n", d)))
                    f.Write([]byte("\n"))
                    timeEpoch := d.ApiData.CompleteTime / 1000
                    if timeEpoch != 0 {
                        duration := time.Unix(timeEpoch, 0).Sub(time.Now())
                        go notifyMissionComplete(duration)
                    }

                default:
                    //f.Write([]byte(jsonBody))
                    f.Write([]byte("\n"))
                }
            }
            return resp
        })
    log.Fatal(ListenAndServe(":8080", proxy))
}
