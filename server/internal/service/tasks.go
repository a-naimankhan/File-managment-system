package service

type ConverTask struct {
	service    *FileService
	InputPath  string
	OutputPath string
}

func (t *ConverTask) Execute() error {
	return t.service.ConvertImageToPDF(t.InputPath, t.OutputPath)
}
