package search

import (
	"errors"
	"io"
	"net/http"
)

func download(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}

	content, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return []byte{}, err
	} else if string(content) == "404: Not Found" {
		return []byte{}, errors.New("404: Not Found")
	}

	return content, nil
}
