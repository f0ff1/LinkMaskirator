package service

type ServiceFactory struct {
	_workers  int
	_slowmode bool //замедление наших воркеров
}

func NewServiceFactory(workers int, slowmode bool) *ServiceFactory {
	return &ServiceFactory{_workers: workers, _slowmode: slowmode}
}

func (f *ServiceFactory) CreateMaskService(inputPath, outputPath string) *Service {
	producer := NewFileProducer(inputPath)
	presenter := NewFilePresenter(outputPath)
	svc := NewService(producer, presenter)
	svc.SetWorkers(f._workers)
	svc.SetSlowMode(f._slowmode)

	return svc
}
