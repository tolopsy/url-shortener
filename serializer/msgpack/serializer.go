package msgpack

import (
	errs "github.com/pkg/errors"
	"github.com/tolopsy/url-shortener/shortener"
	"github.com/vmihailenco/msgpack"
)

type RedirectSerializer struct {}

func (r *RedirectSerializer) Decode(input []byte) (*shortener.Redirect, error) {
	redirect := &shortener.Redirect{}
	if err := msgpack.Unmarshal(input, redirect); err != nil {
		return nil, errs.Wrap(err, "serializer.Redirect.Decode")
	}
	return redirect, nil
}

func (r *RedirectSerializer) Encode(input *shortener.Redirect) ([]byte, error) {
	encodedRedirect, err := msgpack.Marshal(input)
	if err != nil {
		return nil, errs.Wrap(err, "serializer.Redirect.Encode")
	}
	return encodedRedirect, nil
}
