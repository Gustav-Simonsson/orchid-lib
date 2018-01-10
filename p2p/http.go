package p2p

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/log"
)

type HTTPRespHandler func([]byte) ([]byte, error)

func HTTPServer(port int, handler HTTPRespHandler) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error("HTTP REQ", "err", err)
			return
		}
		log.Debug("HTTP REQ", "body", string(b))

		resp, err := handler(b)
		if err != nil {
			log.Error("HTTP handler", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(resp))
	})

	return http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
