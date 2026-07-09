package main

import (
	"flag"
	"log"
	"sync"
)

func main() {
	mode := flag.String("mode", "all", "worker mode: batch|otp|all")
	flag.Parse()

	switch *mode {
	case "batch":
		log.Println("[MainWorker] starting batch worker")
		batchWorker()
	case "otp":
		log.Println("[MainWorker] starting otp worker")
		otpWorker()
	case "all":
		log.Println("[MainWorker] starting batch + otp workers")

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			batchWorker()
		}()

		go func() {
			defer wg.Done()
			otpWorker()
		}()

		wg.Wait()
	default:
		log.Fatalf("invalid mode: %s (supported: batch|otp|all)", *mode)
	}
}
