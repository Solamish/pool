package example

import (
	"fmt"
	"pool"
	"sync"
)

func main() {
    pool, _ := pool.NewPool(5)
	var wg sync.WaitGroup
    a := func() {
			fmt.Println("Hello World!")
			wg.Done()
	}

    for i := 0; i < 10; i++ {
		wg.Add(1)
    	_ =pool.Submit(a)
	}
    wg.Wait()
	fmt.Printf("running goroutines: %d\n", pool.Running())

}


