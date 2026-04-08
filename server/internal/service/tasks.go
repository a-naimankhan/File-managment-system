package service

import "context"

type ConvertTask struct {
	service    *FileService
	InputPath  string
	OutputPath string
}

func (t *ConvertTask) Execute(ctx context.Context) error {
	return t.service.ConvertImageToPDF(t.InputPath, t.OutputPath)
}
