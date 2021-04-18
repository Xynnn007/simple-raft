package http

import (
	"bytes"
	"io/ioutil"

	"net/http"
	oshttp "net/http"

	log "github.com/sirupsen/logrus"
)

type HttpClient struct {
	recvchan chan []byte
}

func (c *HttpClient) Send(address string, port string, data []byte) error {
	reader := bytes.NewReader(data)
	_, err := oshttp.Post("http://"+address+":"+port, "application/json", reader)
	log.Debugf("Post to %v:%v", address, port)

	return err
}

func (c *HttpClient) Recv() chan []byte {
	return c.recvchan
}

func New(port string) *HttpClient {
	h := &HttpClient{
		recvchan: make(chan []byte),
	}

	oshttp.HandleFunc("/", h.ServeHTTP)
	go func() {
		log.Infof("Http start serving at port %v...", port)
		oshttp.ListenAndServe(":"+port, nil)
	}()
	return h
}

func (c *HttpClient) ServeHTTP(_ http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Http get request failed: %v", err)
		return
	}
	log.Debugf("Got message from %v", r.RemoteAddr)
	c.recvchan <- body
}
