package utils_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/wythers/algalon/utils"
	"gopkg.in/go-playground/assert.v1"
)

type placeholder struct{}

func Test_queue(t *testing.T) {
	queue := utils.NewQueue[placeholder]()

	ps := make([]*placeholder, 0, 32)
	for i := 0; i < 1100; i++ {
		ps = append(ps, &placeholder{})
	}

	queue.BatchIn(ps)
	assert.Equal(t, 1100, queue.Counter())

	wg := sync.WaitGroup{}
	flag := int32(0)
	wg.Add(19000)

	// consumer
	for i := 0; i < 10000; i++ {
		go func() {
			defer wg.Done()
			for {
				if atomic.LoadInt32(&flag) == 0 {
					continue
				}
				break
			}

			for {
				_, err := queue.Dequeue()
				if err != nil {
					continue
				}

				break
			}
		}()
	}

	for i := 0; i < 9000; i++ {
		go func() {
			defer wg.Done()
			for {
				if atomic.LoadInt32(&flag) == 0 {
					continue
				}
				break
			}

			queue.Enqueue(&placeholder{})
		}()
	}

	atomic.AddInt32(&flag, 1)
	wg.Wait()

	assert.Equal(t, 100, queue.Counter())
	assert.Equal(t, false, queue.IsEmpty())

	tmp, _ := queue.BatchOut()

	assert.Equal(t, 100, len(tmp))
	assert.Equal(t, 0, queue.Counter())
	assert.Equal(t, true, queue.IsEmpty())

}
