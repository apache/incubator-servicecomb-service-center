//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package store

import (
	"errors"
	"github.com/ServiceComb/service-center/pkg/util"
	"golang.org/x/net/context"
	"sync"
	"time"
)

const (
	DEFAULT_MAX_TASK_COUNT        = 1000
	DEFAULT_REMOVE_TASKS_INTERVAL = 30 * time.Second
)

type AsyncTask interface {
	Key() string
	Do(ctx context.Context) error
	Err() error
}

type schedule struct {
	queue      *util.UniQueue
	latestTask AsyncTask
}

func (s *schedule) AddTask(ctx context.Context, task AsyncTask) error {
	if task == nil || ctx == nil {
		return errors.New("invalid parameters")
	}

	err := s.queue.Put(ctx, task)
	if err != nil {
		return err
	}
	util.Logger().Debugf("add task done! key is %s", task.Key())
	return s.latestTask.Err()
}

func (s *schedule) collectReadyTasks(ctx context.Context, ready chan<- AsyncTask) {
	defer util.RecoverAndReport()
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-s.queue.Chan():
			if !ok {
				return
			}
			ready <- task
		}
	}
}

type AsyncTasker struct {
	schedules   map[string]*schedule
	queues      map[string]*util.UniQueue
	latestTasks map[string]AsyncTask
	removeTasks map[string]struct{}
	goroutine   *util.GoRoutine
	queueLock   sync.RWMutex
	ready       chan struct{}
	isClose     bool
}

func (lat *AsyncTasker) getOrNewSchedule(key string, task AsyncTask) (s *schedule, isNew bool) {
	var ok bool

	lat.queueLock.RLock()
	s, ok = lat.schedules[key]
	_, remove := lat.removeTasks[key]
	lat.queueLock.RUnlock()
	if !ok {
		lat.queueLock.Lock()
		s, ok = lat.schedules[key]
		if !ok {
			isNew = true
			lat.schedules[key] = &schedule{
				queue: util.NewUniQueue(),
			}
			lat.goroutine.Do(func() {})
		}
		lat.queueLock.Unlock()
	}
	if remove && ok {
		lat.queueLock.Lock()
		_, remove = lat.removeTasks[key]
		if remove {
			delete(lat.removeTasks, key)
		}
		lat.queueLock.Unlock()
	}
	return
}

func (lat *AsyncTasker) getOrNewQueue(key string, task AsyncTask) (*util.UniQueue, bool) {
	lat.queueLock.RLock()
	queue, ok := lat.queues[key]
	_, remove := lat.removeTasks[key]
	lat.queueLock.RUnlock()
	if !ok {
		lat.queueLock.Lock()
		queue, ok = lat.queues[key]
		if !ok {
			queue = util.NewUniQueue()
			lat.queues[key] = queue
			lat.latestTasks[key] = task
		}
		lat.queueLock.Unlock()
	}
	if remove && ok {
		lat.queueLock.Lock()
		_, remove = lat.removeTasks[key]
		if remove {
			delete(lat.removeTasks, key)
		}
		lat.queueLock.Unlock()
	}
	return queue, !ok
}

func (lat *AsyncTasker) AddTask(ctx context.Context, task AsyncTask) error {
	if task == nil || ctx == nil {
		return errors.New("invalid parameters")
	}

	s, isNew := lat.getOrNewSchedule(task.Key(), task)
	if isNew {
		s.latestTask = task
		// do immediately at first time
		return task.Do(ctx)
	}
	return s.AddTask(ctx, task)
}

func (lat *AsyncTasker) DeferRemoveTask(key string) error {
	lat.queueLock.Lock()
	if lat.isClose {
		lat.queueLock.Unlock()
		return errors.New("AsyncTasker is stopped")
	}
	_, exist := lat.queues[key]
	if !exist {
		lat.queueLock.Unlock()
		return nil
	}
	lat.removeTasks[key] = struct{}{}
	lat.queueLock.Unlock()
	return nil
}

func (lat *AsyncTasker) removeTask(key string) {
	if queue, ok := lat.queues[key]; ok {
		queue.Close()
		delete(lat.queues, key)
	}
	delete(lat.latestTasks, key)
	delete(lat.removeTasks, key)
	util.Logger().Debugf("remove task, key is %s", key)
}

func (lat *AsyncTasker) LatestHandled(key string) (AsyncTask, error) {
	lat.queueLock.RLock()
	at, ok := lat.latestTasks[key]
	lat.queueLock.RUnlock()
	if !ok {
		return nil, errors.New("expired behavior")
	}
	return at, nil
}

func (lat *AsyncTasker) schedule(stopCh <-chan struct{}) {
	util.SafeCloseChan(lat.ready)

	tasks := make(chan AsyncTask, DEFAULT_MAX_TASK_COUNT)
	defer func() {
		close(tasks)
		util.Logger().Debugf("AsyncTasker is ready to stop")
	}()

	go lat.collectReadyTasks(stopCh, tasks)

	for {
		select {
		case <-stopCh:
			util.Logger().Debugf("scheduler exited for AsyncTasker is stopped")
			return
		case task := <-tasks:
			lat.scheduleTask(task.(AsyncTask))
		}
	}
}

func (lat *AsyncTasker) collectReadyTasks(stopCh <-chan struct{}, tasks chan<- AsyncTask) {
	defer util.RecoverAndReport()

	if len(lat.queues) == 0 {
		return
	}

	lat.queueLock.RLock()
	defer lat.queueLock.RUnlock()

	for key, queue := range lat.queues {
		select {
		case task, ok := <-queue.Chan():
			if !ok {
				util.Logger().Warnf(nil, "get a closed task in queue and discard it, key is %s", key)
				continue
			}
			tasks <- task.(AsyncTask) // will block when a lot of tasks coming in.
		default:
		}
	}
}

func (lat *AsyncTasker) daemonRemoveTask(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			util.Logger().Debugf("daemon remove task exited for AsyncTasker is stopped")
			return
		case <-time.After(DEFAULT_REMOVE_TASKS_INTERVAL):
			if lat.isClose {
				return
			}
			lat.queueLock.Lock()
			l := len(lat.removeTasks)
			for key := range lat.removeTasks {
				lat.removeTask(key)
			}
			lat.queueLock.Unlock()
			if l > 0 {
				util.Logger().Infof("daemon remove task completed, %d removed", l)
			}
		}
	}
}

func (lat *AsyncTasker) Run() {
	lat.queueLock.Lock()
	if !lat.isClose {
		lat.queueLock.Unlock()
		return
	}
	lat.isClose = false
	lat.queueLock.Unlock()
	lat.goroutine.Do(lat.schedule)
	lat.goroutine.Do(lat.daemonRemoveTask)
}

func (lat *AsyncTasker) scheduleReadyTasks(ready <-chan AsyncTask) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	for {
		select {
		case task := <-ready:
			lat.scheduleTask(task.(AsyncTask))
		case <-ctx.Done():
			return
		}
	}
}

func (lat *AsyncTasker) scheduleTask(at AsyncTask) {
	util.Logger().Debugf("start to run task, key is %s", at.Key())
	lat.goroutine.Do(func(stopCh <-chan struct{}) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			defer cancel()

			lat.queueLock.RLock()
			_, ok := lat.latestTasks[at.Key()]
			lat.queueLock.RUnlock()
			if !ok {
				util.Logger().Warnf(nil, "can not schedule a closed task, key is %s", at.Key())
				return
			}

			at.Do(ctx)

			lat.UpdateLatestTask(at)
		}()
		select {
		case <-ctx.Done():
			util.Logger().Debugf("finish to handle task, key is %s, result: %s", at.Key(), at.Err())
		case <-stopCh:
			cancel()
			util.Logger().Debugf("cancelled task for AsyncTasker is stopped, key is %s", at.Key())
		}
	})
}

func (lat *AsyncTasker) UpdateLatestTask(at AsyncTask) {
	lat.queueLock.RLock()
	_, ok := lat.latestTasks[at.Key()]
	lat.queueLock.RUnlock()
	if !ok {
		return
	}

	lat.queueLock.Lock()
	_, ok = lat.latestTasks[at.Key()]
	if ok {
		lat.latestTasks[at.Key()] = at
	}
	lat.queueLock.Unlock()
}

func (lat *AsyncTasker) Stop() {
	lat.queueLock.Lock()
	if lat.isClose {
		lat.queueLock.Unlock()
		return
	}
	lat.isClose = true
	for key := range lat.queues {
		lat.removeTask(key)
	}
	lat.queueLock.Unlock()
	lat.goroutine.Close(true)

	util.SafeCloseChan(lat.ready)

	util.Logger().Debugf("AsyncTasker is stopped")
}

func (lat *AsyncTasker) Ready() <-chan struct{} {
	return lat.ready
}

func NewAsyncTasker() *AsyncTasker {
	return &AsyncTasker{
		latestTasks: make(map[string]AsyncTask),
		queues:      make(map[string]*util.UniQueue),
		removeTasks: make(map[string]struct{}),
		goroutine:   util.NewGo(make(chan struct{})),
		ready:       make(chan struct{}),
		isClose:     true,
	}
}
