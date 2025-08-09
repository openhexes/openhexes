package progress

import (
	"context"

	"github.com/openhexes/openhexes/api/src/config"
	progressv1 "github.com/openhexes/proto/progress/v1"
	"go.uber.org/zap"
)

type SendFunc func(*progressv1.Progress) error

type Reporter struct {
	ctx  context.Context
	send SendFunc
	msg  *progressv1.Progress
	ch   chan struct{}
	done chan struct{}
}

func NewReporter(ctx context.Context, send SendFunc, stages ...*progressv1.Stage) *Reporter {
	r := &Reporter{
		ctx:  ctx,
		send: send,
		msg: &progressv1.Progress{
			Stages: stages,
		},
		ch:   make(chan struct{}, 20),
		done: make(chan struct{}),
	}

	go r.run()

	return r
}

func (r *Reporter) Update(percentage ...float64) {
	if len(percentage) > 0 {
		r.msg.Percentage = percentage[len(percentage)-1]
	}
	r.ch <- struct{}{}
}

func (r *Reporter) Close() {
	close(r.ch)
	<-r.done
}

func (r *Reporter) run() {
	log := config.GetLogger(r.ctx)

	for range r.ch {
		if err := r.send(r.msg); err != nil {
			log.Warn("failed to send progress", zap.Error(err))
		}
	}

	close(r.done)

	// Loop:
	//
	//	for {
	//		select {
	//		case <-r.ctx.Done():
	//			break Loop
	//		case _, ok := <-r.ch:
	//			if !ok {
	//				break Loop
	//			}
	//			if err := r.send(r.msg); err != nil {
	//				log.Warn("failed to send progress", zap.Error(err))
	//			}
	//		}
	//	}
}
