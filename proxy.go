package main

import (
    "bufio"
    "bytes"
    "github.com/elazarl/goproxy"
    "io/ioutil"
    "log"
    ."net/http"
    "os"
    "regexp"
)

const (
    LOG_FILENAME = "log.txt"
)

func main() {
    proxy := goproxy.NewProxyHttpServer()
    proxy.Verbose = true
    f, err := os.Create(LOG_FILENAME)
    if err != nil {
        panic(err)
    }

    re := regexp.MustCompile(`kcsapi`)
    proxy.OnResponse().DoFunc(
        func (resp *Response, ctx *goproxy.ProxyCtx) *Response {
            r := resp.Request
            if re.MatchString(r.URL.Path) == true {
                data := []byte(r.URL.Path + "\n")
                f.Write(data)
                reader := bufio.NewReader(resp.Body)
                b, _ := ioutil.ReadAll(reader)
                buf := bytes.NewBuffer(b)
                resp.Body = ioutil.NopCloser(buf)
                f.Write(b)
                f.Write([]byte("\n"))
            }
            return resp
    })
    log.Fatal(ListenAndServe(":8080", proxy))
}

