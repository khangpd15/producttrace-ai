package qr

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"
)

type PDFGenerator interface {
	GenerateLabels(input BatchPDFInput) ([]byte, error)
}

type pdfGenerator struct {
	qr      Generator
	baseURL string
}

func NewPDFGenerator(qr Generator, baseURL string) PDFGenerator {
	return &pdfGenerator{
		qr:      qr,
		baseURL: baseURL,
	}
}

func (p *pdfGenerator) GenerateLabels(input BatchPDFInput) ([]byte, error) {

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	title := fmt.Sprintf("BATCH: %s", input.BatchCode)
	pdf.SetFont("Helvetica", "B", 14)

	pdf.SetY(10)

	pdf.CellFormat(
		0,
		10,
		title,
		"",
		1,
		"C",
		false,
		0,
		"",
	)

	pdf.Ln(5)

	const (
		qrSize = 35.0
		cellW  = 60.0
		cellH  = 50.0

		ItemCodeW = 35.0

		startX = 25.0
		startY = 30.0

		cols = 3
	)

	x := startX
	y := startY

	for i, item := range input.Items {

		verifyURL := fmt.Sprintf(
			"https://your-frontend-domain.vercel.app/verify?item_code=%s&token=%s",
			item.ItemCode,
			item.Token,
		)

		qrBytes, err := p.qr.Generate(
			verifyURL,
		)
		if err != nil {
			return nil, err
		}

		imageName := fmt.Sprintf(
			"qr-%d",
			i,
		)

		opts := fpdf.ImageOptions{
			ImageType: "PNG",
		}

		pdf.RegisterImageOptionsReader(
			imageName,
			opts,
			bytes.NewReader(qrBytes),
		)

		pdf.ImageOptions(
			imageName,
			x,
			y,
			qrSize,
			qrSize,
			false,
			opts,
			0,
			"",
		)

		pdf.SetXY(
			x,
			y+qrSize+2,
		)

		pdf.SetFont(
			"Helvetica",
			"",
			8,
		)

		pdf.CellFormat(
			ItemCodeW,
			4,
			item.ItemCode,
			"",
			0,
			"C",
			false,
			0,
			"",
		)

		col := (i + 1) % cols

		if col == 0 {
			x = startX
			y += cellH
		} else {
			x += cellW
		}

		if y+cellH > 280 {
			pdf.AddPage()
			x = startX
			y = startY
		}
	}

	var buf bytes.Buffer

	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
