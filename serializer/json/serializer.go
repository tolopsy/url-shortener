package json

import (
	"encoding/json"
	errs "github.com/pkg/errors"
	"github.com/tolopsy/url-shortener/shortener"
)

type RedirectSerializer struct {}

func (r *RedirectSerializer) Decode(input []byte) (*shortener.Redirect, error) {
	redirect := &shortener.Redirect{}
	if err := json.Unmarshal(input, redirect); err != nil {
		return nil, errs.Wrap(err, "serializer.Redirect.Decode")
	}
	return redirect, nil
}

func (r *RedirectSerializer) Encode(input *shortener.Redirect) ([]byte, error) {
	encodedRedirect, err := json.Marshal(input)
	if err != nil {
		return nil, errs.Wrap(err, "serializer.Redirect.Encode")
	}
	return encodedRedirect, nil
}
