package parse

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"
)

type CSVParseOptions struct {
	Delimiter       rune
	Comment         rune
	FieldsPerRecord int
	ReadFirstLine   bool
}

func ParseCSV(data []byte, options CSVParseOptions) ([][]string, error) {
	reader := csv.NewReader(bytes.NewReader(data))

	if options.Delimiter != 0 {
		reader.Comma = options.Delimiter
	}

	if options.Comment != 0 {
		reader.Comment = options.Comment
	}

	if options.FieldsPerRecord != 0 {
		reader.FieldsPerRecord = options.FieldsPerRecord
	}

	var records [][]string
	record, eof, err := processCSVLine(reader)
	if err != nil || eof {
		return nil, err
	}

	if options.ReadFirstLine {
		records = append(records, record)
	}

	for {
		record, eof, err := processCSVLine(reader)

		if eof {
			return records, nil
		}

		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}
}

func processCSVLine(reader *csv.Reader) ([]string, bool, error) {
	record, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, true, nil
		}
		return nil, false, err
	}

	for i, field := range record {
		record[i] = strings.Trim(field, " ")
	}

	return record, false, nil
}
