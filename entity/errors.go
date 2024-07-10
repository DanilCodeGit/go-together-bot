package entity

import "github.com/pkg/errors"

var (
	ErrStatusCode  = errors.New("status code error")
	ErrSendSticker = errors.New("send sticker")
	ErrResponse    = errors.New("response error")
	ErrSendMsg     = errors.New("send message error")
)
