package qr

import "github.com/skip2/go-qrcode"

type Generator interface {
	Generate(content string) ([]byte, error)
}

type generator struct {
	size int
}

func NewGenerator() Generator {
	return &generator{
		size: 256,
	}
}

func (g *generator) Generate(content string) ([]byte, error) {
	return qrcode.Encode(
		content,
		qrcode.Medium,
		g.size,
	)
}
