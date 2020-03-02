package main

import (
	"fmt"
	"time"
)

/*
	发布订阅模式是一种消息通知模式，发布者发送消息，订阅者接收消息。
 */

type Subscriber struct {
	in     chan interface{}
	id     int
	topic  string
	stop   chan struct{}
}

func (s *Subscriber) Close() {
	s.stop <- struct{}{}
	close(s.in)
}

func (s *Subscriber) Notify(msg interface{}) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("%#v", rec)
		}
	}()
	select {
	case s.in <-msg:
	case <-time.After(time.Second):
		err = fmt.Errorf("Timeout\n")
	}
	return
}

func NewSubscriber(id int) SubscriberImpl {
	s := &Subscriber{
		id: id,
		in: make(chan interface{}),
		stop: make(chan struct{}),
	}
	go func() {
		for{
			select {
			case <-s.stop:
				close(s.stop)
				return
			default:
				for msg := range s.in {
					fmt.Printf("(W%d): %v\n", s.id, msg)
				}
			}
		}}()
	return s
}

// 订阅者需要实现的方法
type SubscriberImpl interface {
	Notify(interface{}) error
	Close()
}

// sub 订阅 pub
func Register(sub Subscriber, pub *publisher){
	pub.addSubCh <- sub
	return
}

// pub 结果定义
type publisher struct {
	subscribers []SubscriberImpl
	addSubCh    chan SubscriberImpl
	removeSubCh chan SubscriberImpl
	in          chan interface{}
	stop        chan struct{}
}

// 实例化
func NewPublisher () *publisher{
	return &publisher{
		addSubCh: make(chan SubscriberImpl),
		removeSubCh: make(chan SubscriberImpl),
		in: make(chan interface{}),
		stop: make(chan struct{}),
	}
}

// 监听
func (p *publisher) start() {
	for {
		select {
		// pub 发送消息
		case msg := <-p.in:
			for _, sub := range p.subscribers{
				_ = sub.Notify(msg)
			}
		// 移除指定 sub
		case sub := <-p.removeSubCh:
			for i, candidate := range p.subscribers {
				if candidate == sub {
					p.subscribers = append(p.subscribers[:i], p.subscribers[i+1:]...)
					candidate.Close()
					break
				}
			}
		// 增加一个 sub
		case sub := <-p.addSubCh:
			p.subscribers = append(p.subscribers, sub)
		// 关闭 pub
		case <-p.stop:
			for _, sub := range p.subscribers {
				sub.Close()
			}
			close(p.addSubCh)
			close(p.in)
			close(p.removeSubCh)
			return
		}
	}
}

func main() {
	// 测试代码
	pub := NewPublisher()
	go pub.start()

	sub1 := NewSubscriber(1)
	Register(sub1, pub)

	sub2 := NewSubscriber(2)
	Register(sub2, pub)

	commands:= []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for _, c := range commands {
		pub.in <- c
	}

	pub.stop <- struct{}{}
	time.Sleep(time.Second*1)
}