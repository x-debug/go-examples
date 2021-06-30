package timex

import (
	"container/list"
	"errors"
	"log"
	"time"
)

type task struct {
	key    string
	circle int
	pos    int
	fn     func()
	slot   *list.List
}

type setingTasker struct {
	key      string
	duration time.Duration
	callback func()
}

type removingTasker struct {
	key string
}

type TimeWheel struct {
	tickerPos  int
	interval   int
	sizeOfSlot int
	wheel      []*list.List
	wMap       map[string]*list.Element
	setChan    chan *setingTasker
	removeChan chan *removingTasker
	ticker     *time.Ticker
}

func NewTimeWheel(interval int, sizeOfSlot int) *TimeWheel {
	wheel := &TimeWheel{
		interval:   interval,
		tickerPos:  0,
		sizeOfSlot: sizeOfSlot,
		ticker:     time.NewTicker(time.Duration(interval) * time.Second),
		setChan:    make(chan *setingTasker),
		removeChan: make(chan *removingTasker),
		wMap:       make(map[string]*list.Element)}
	wheel.initWheel(wheel.sizeOfSlot)
	go wheel.run()
	return wheel
}

func (tw *TimeWheel) initWheel(slotNumber int) {
	tw.wheel = make([]*list.List, tw.sizeOfSlot)
	for i := 0; i < tw.sizeOfSlot; i++ {
		tw.wheel[i] = list.New()
	}
}

func (tw *TimeWheel) run() {
	for {
		select {
		case tasker := <-tw.setChan:
			tw.setTime(tasker)
		case tasker := <-tw.removeChan:
			tw.removeTime(tasker)
		case <-tw.ticker.C:
			tw.runTicker()
		}
	}
}

func (tw *TimeWheel) SetTimer(key string, d time.Duration, f func()) error {
	if _, ok := tw.wMap[key]; ok {
		return errors.New("timer's key is exist")
	}

	tasker := &setingTasker{key: key, duration: d, callback: f}
	tw.setChan <- tasker
	return nil
}

func (tw *TimeWheel) RemoveTimer(key string) error {
	if _, ok := tw.wMap[key]; !ok {
		return errors.New("timer's key is not exist")
	}

	tasker := &removingTasker{key: key}
	tw.removeChan <- tasker
	return nil
}

func (tw *TimeWheel) StopTimer() {
	tw.ticker.Stop()
}

func (tw *TimeWheel) getPosition(d time.Duration) (circle int, pos int) {
	circle = (int(d) / int(time.Second) / tw.interval) / tw.sizeOfSlot
	pos = (tw.tickerPos + int(d)/int(time.Second)/tw.interval) % tw.sizeOfSlot
	return
}

func (tw *TimeWheel) setTime(tasker *setingTasker) {
	key, d, fn := tasker.key, tasker.duration, tasker.callback
	//TODO update timer
	if int(d) < tw.interval {
		d = time.Duration(tw.interval)
	}
	circle, pos := tw.getPosition(d)
	slot := tw.wheel[pos]
	tw.insertSlot(slot, key, &task{circle: circle, pos: pos, fn: fn, slot: slot, key: key})
}

func (tw *TimeWheel) insertSlot(slot *list.List, key string, tasker *task) *list.Element {
	element := slot.PushBack(tasker)
	tw.wMap[key] = element
	return element
}

func (tw *TimeWheel) removeTime(rmTasker *removingTasker) {
	if element, ok := tw.wMap[rmTasker.key]; ok {
		tasker := element.Value.(*task)
		tasker.slot.Remove(element)
		delete(tw.wMap, rmTasker.key)
	}
}

func (tw *TimeWheel) runTicker() {
	tw.tickerPos = (tw.tickerPos + 1) % tw.sizeOfSlot
	slot := tw.wheel[tw.tickerPos]
	for e := slot.Front(); e != nil; e = e.Next() {
		tasker := e.Value.(*task)
		if tasker.circle > 0 {
			tasker.circle -= 1
			continue
		} else {
			tw.removeTime(&removingTasker{key: tasker.key})
			go func() {
				defer func() {
					if err := recover(); err != nil {
						log.Printf("tasker error: %v\n", err)
					}
				}()

				tasker.fn()
			}()
		}
	}
}
