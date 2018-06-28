package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/golang/glog"
	"github.com/satori/go.uuid"
)

var (
	name = flag.String("name", "thumb.web", "service name")
	host = flag.String("host", ":9066", "listen ip and port")
)

func GetQuery(r *http.Request, key string, defaultVal string) string {
	values, ok := r.URL.Query()[key]
	if ok && len(values) > 0 && len(values[0]) > 0 {
		return values[0]
	}
	return defaultVal
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		glog.Infoln(r.Method, r.URL, r.RemoteAddr)
		handler.ServeHTTP(w, r)
	})
}

func echoFunc(w http.ResponseWriter, r *http.Request) {
	echoStr := r.URL.Query().Get("str")
	if echoStr == "" {
		echoStr = "ok"
	}
	fmt.Fprintf(w, echoStr)
}

func thumbFunc(w http.ResponseWriter, r *http.Request) {
	imageId := uuid.NewV4().String()
	imageFile := fmt.Sprintf("/tmp/%s.png", imageId)

	url := GetQuery(r, "url", "http://www.baidu.com")
	if false == strings.HasPrefix(url, "http://") && false == strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	width := GetQuery(r, "width", "1024")
	height := GetQuery(r, "height", "768")
	cmdStr := fmt.Sprintf("xvfb-run --server-args=\"-screen 0, 1920x1080x24\" CutyCapt --url=\"%s\" --min-width=%s --min-height=%s --out=%s", url, width, height, imageFile)

	glog.Infoln("execute command : ", cmdStr)
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		glog.Infoln("execute command failed, err = ", err)
		w.Write([]byte("execute command failed, err = " + err.Error()))
		return
	}

	contents, err := ioutil.ReadFile(imageFile)
	w.Header().Set("content-type", "image/png")
	w.Write(contents)

	os.Remove(imageFile)
}

func main() {
	flag.Parse()

	http.HandleFunc("/echo", echoFunc)
	http.HandleFunc("/", thumbFunc)

	if err := http.ListenAndServe(*host, Log(http.DefaultServeMux)); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}