package object

import "container/list"

type Queue struct {
	// TODO: using array/slice to implement for performace
	list *list.List
}

func NewQueue() *Queue {
	return &Queue{
		list: list.New(),
	}
}

func (q *Queue) Enqueue(t *Tree) {
	q.list.PushBack(t)
}

func (q *Queue) Dequeue() *Tree {
	t := q.list.Front()

	if t != nil {
		q.list.Remove(t)
		return t.Value.(*Tree)
	}
	return nil
}
