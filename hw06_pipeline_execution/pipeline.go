package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	pipeOut := make(Bi)
	pipeIn := make(Bi)
	var stageOut Out = pipeIn

	for _, stage := range stages {
		stageOut = stage(stageOut)
	}

	go func() {
		defer close(pipeIn)
		for v := range in {
			select {
			case <-done:
				return
			case pipeIn <- v:
			}
		}
	}()

	go func() {
		defer func() {
			close(pipeOut)
			// нужно вычитывать stageOut, чтобы цепочка каналов пайплайна не подвисла из-за отсутствия читателя
			// читатель может пропасть, если пришёл сигнал done, а пайплайн уже запущен
			for v := range stageOut {
				_ = v
			}
		}()
		for {
			select {
			case <-done:
				return
			case v, more := <-stageOut:
				if more {
					pipeOut <- v
				} else {
					return
				}
			}
		}
	}()

	return pipeOut
}
