package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type fork struct {
	name string
	mu sync.Mutex
}

func NewFork(name string) *fork {
	var mu sync.Mutex
	return &fork{
		name: name,
		mu: mu,
	}
}

func (f *fork) Take() { f.mu.Lock() }
func (f *fork) Put() { f.mu.Unlock() }
func (f *fork) Name() string { return f.name}

type philosopher struct {
	name string
	left *fork
	right *fork
}

func NewPhilosopher(name string, left, right *fork) *philosopher {
	return &philosopher{
		name: name,
		left: left,
		right: right,
	}
}

func (p philosopher) Eat(ctx context.Context, waitTurn chan struct{}) {
LOOP:
	for {
		select {
		case waitTurn <- struct{}{}:
			p.take(p.left)
			p.take(p.right)
			fmt.Printf("%s is eating\n", p.name)
			time.Sleep(time.Second)
			p.put(p.right)
			p.put(p.left)
			select {
			case <-waitTurn:
			case <-ctx.Done():
				break LOOP
			}
		case <-ctx.Done():
			break LOOP
		}
		fmt.Printf("%s is thinking...\n", p.name)
		time.Sleep(time.Second)
	}
}

func (p philosopher) take(f *fork) {
	f.Take()
	fmt.Printf("%s took %s\n", p.name, f.Name())
}

func (p philosopher) put(f *fork) {
	f.Put()
	fmt.Printf("%s put %s\n", p.name, f.Name())
}

const num = 5

func main() {
	forks := make([]*fork, num)
	philosophers := make([]*philosopher, num)
	sem := make(chan struct{}, num - 1)
	for i := 0; i < num; i++ {
		forks[i] = NewFork(fmt.Sprintf("fork %d", i + 1))
	}
	for i := 0; i < num; i++ {
		j := i - 1
		if i == 0 {
			j = num - 1
		}
		philosophers[i] = NewPhilosopher(fmt.Sprintf("philosopher %d", i + 1), forks[i], forks[j])
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(philosophers))
	for _, p := range philosophers {
		go func(p *philosopher) {
			p.Eat(ctx, sem)
			wg.Done()
		}(p)
	}
	wg.Wait()
}
