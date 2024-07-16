package parse

import "testing"

func TestParseCSV(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		options CSVParseOptions
		want    [][]string
	}{
		{
			name:    "nominal case",
			data:    []byte("Name,Age\nJohn,Smith\nMarc,Doe"),
			options: CSVParseOptions{},
			want:    [][]string{{"John", "Smith"}, {"Marc", "Doe"}},
		},
		{
			name:    "empty data",
			data:    []byte{},
			options: CSVParseOptions{},
			want:    [][]string{},
		},
		{
			name:    "extra spaces",
			data:    []byte("  Name ,  Last Name \n John   , Smith    \n Jean Marc  , Doe   "),
			options: CSVParseOptions{},
			want:    [][]string{{"John", "Smith"}, {"Jean Marc", "Doe"}},
		},
		{
			name:    "delimiter option",
			data:    []byte("Name;Lase Name\nJohn;Smith\nMarc;Doe"),
			options: CSVParseOptions{Delimiter: ';'},
			want:    [][]string{{"John", "Smith"}, {"Marc", "Doe"}},
		},
		{
			name:    "comment option",
			data:    []byte("Name, Last Name\n# some, comment\nJohn,Smith\nMarc,Doe"),
			options: CSVParseOptions{Comment: '#'},
			want:    [][]string{{"John", "Smith"}, {"Marc", "Doe"}},
		},
		{
			name:    "read first line option",
			data:    []byte("Name,Last Name\nJohn,Smith\nMarc,Doe"),
			options: CSVParseOptions{ReadFirstLine: true},
			want:    [][]string{{"Name", "Last Name"}, {"John", "Smith"}, {"Marc", "Doe"}},
		},
		{
			name:    "fields per record positive option",
			data:    []byte("Name,Age\nJohn,Smith\nMarc,Doe"),
			options: CSVParseOptions{FieldsPerRecord: 2},
			want:    [][]string{{"John", "Smith"}, {"Marc", "Doe"}},
		},
		{
			name:    "fields per record negative option",
			data:    []byte("Name,Age\nJohn,Smith\nMarc"),
			options: CSVParseOptions{FieldsPerRecord: -1},
			want:    [][]string{{"John", "Smith"}, {"Marc"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCSV(tt.data, tt.options)
			if err != nil {
				t.Error(err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("ParseCSV() = %v, want %v", got, tt.want)
				return
			}

			for i := range got {
				if len(got[i]) != len(tt.want[i]) {
					t.Errorf("ParseCSV() = %v, want %v", got, tt.want)
					continue
				}

				for j := range got[i] {
					if got[i][j] != tt.want[i][j] {
						t.Errorf("ParseCSV() = %v, want %v", got, tt.want)
					}
				}
			}
		})
	}
}
