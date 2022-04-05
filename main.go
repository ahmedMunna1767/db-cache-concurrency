package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var cache = map[int]Book{}
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	wg := &sync.WaitGroup{}
	m := &sync.RWMutex{}

	cacheCh := make(chan Book)
	dbCh := make(chan Book)
	for i := 0; i < 10; i++ {
		id := rnd.Intn(10) + 1
		wg.Add(2)
		go func(id int, wg *sync.WaitGroup, m *sync.RWMutex, ch chan<- Book) {
			// Query the cache, if found then print it out
			if b, ok := queryFromCache(id, m); ok {
				ch <- b // Pass the found Book to the cache channel
			}

			wg.Done()
		}(id, wg, m, cacheCh)

		go func(id int, wg *sync.WaitGroup, m *sync.RWMutex, ch chan<- Book) {
			// Query the DB, if found then print it out
			if b, ok := queryFromDB(id, m); ok {
				ch <- b
			}

			wg.Done()
		}(id, wg, m, dbCh)

		go func(cacheCh, dbCh <-chan Book) {
			select {
			case b := <-cacheCh:
				fmt.Println("Source: Cache")
				fmt.Println(b.String())
				<-dbCh
				// fmt.Println(<-dbCh) // Wait to get the message from the DB Channel so we don't block
			case b := <-dbCh:
				fmt.Println("Source: Database")
				fmt.Println(b.String())
			}
		}(cacheCh, dbCh)

		time.Sleep(150 * time.Millisecond)
	}
	wg.Wait()
}

func queryFromDB(id int, m *sync.RWMutex) (Book, bool) {
	time.Sleep(time.Duration(time.Millisecond * 150))
	for _, book := range books {
		if book.ID == id {
			m.Lock()
			cache[id] = book
			m.Unlock()
			return book, true
		}
	}
	return Book{}, false
}

func queryFromCache(id int, m *sync.RWMutex) (Book, bool) {
	m.RLock()
	book, ok := cache[id]
	m.RUnlock()
	return book, ok
}
