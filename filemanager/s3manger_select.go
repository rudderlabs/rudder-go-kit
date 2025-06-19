package filemanager

type (
	SelectObjectInputFormat  string
	SelectObjectOutputFormat string
)

const (
	// input formats
	SelectObjectInputFormatParquet SelectObjectInputFormat = "parquet"

	// output formats
	SelectObjectOutputFormatCSV  SelectObjectOutputFormat = "csv"
	SelectObjectOutputFormatJSON SelectObjectOutputFormat = "json"
)

type SelectConfig struct {
	SQLExpression string
	Key           string
	InputFormat   SelectObjectInputFormat
	OutputFormat  SelectObjectOutputFormat
}

type SelectResult struct {
	Data  []byte
	Error error
}
