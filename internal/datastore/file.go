package datastore

import (
	"encoding/csv"
	"os"
	"reflect"
	"strconv"
)

type File struct {
	Name string
	Path string
	Type string
}

type Writer interface {
	Write(data any) error
	FullPath() string
}

type CSVWriter struct {
	File    File
	Headers []string
	Offset  int
}

func NewCSVWriter(file File, headers []string) *CSVWriter {
	return &CSVWriter{File: file, Headers: headers}
}

func (w *CSVWriter) FullPath() string {
	return w.File.Path + "/" + w.File.Name + "." + w.File.Type
}

func (w *CSVWriter) Write(data any) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return nil
	}

	f, err := os.OpenFile(w.FullPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)

	// Write headers if new file
	info, err := f.Stat()
	if err == nil && info.Size() == 0 && len(w.Headers) > 0 {
		writer.Write(w.Headers)
	}

	for i := 0; i < v.Len(); i++ {
		row := v.Index(i)
		fields := extractFields(row.Interface())
		writer.Write(fields)
		w.Offset++
	}
	writer.Flush()
	return writer.Error()
}

func extractFields(obj any) []string {
	v := reflect.ValueOf(obj)
	var row []string

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		switch field.Kind() {
		case reflect.String:
			row = append(row, field.String())
		case reflect.Float64:
			row = append(row, strconv.FormatFloat(field.Float(), 'f', 2, 64))
		case reflect.Int, reflect.Int64:
			row = append(row, strconv.FormatInt(field.Int(), 10))
		default:
			row = append(row, "")
		}
	}
	return row
}