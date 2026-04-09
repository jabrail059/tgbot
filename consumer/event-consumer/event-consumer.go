package event_consumer

import (
	"log"
	"sync"
	"telegrambot/events"
	"time"
)

type Consumer struct {
	fetcher      events.Fetcher
	processor    events.Processor
	batchSize    int
	workersCount int
}

func New(fetcher events.Fetcher, processor events.Processor, batchSize int) Consumer {
	return Consumer{
		fetcher:      fetcher,
		processor:    processor,
		batchSize:    batchSize,
		workersCount: 5,
	}
}

func (c Consumer) Start() error {
	for {
		gotEvents, err := c.fetcher.Fetch(c.batchSize)
		if err != nil {
			log.Printf("[ERR] consumer: %s", err.Error())
			continue
		}

		if len(gotEvents) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}
		if err := c.handleEvents(gotEvents); err != nil {
			log.Print(err)
			continue
		}
	}
}

/*
1. Потеря событий: ретраи, возвращение в хранилище, фоллбэк, подтверждение для fetcher
2. Обработка всей пачки: останавливаться после первой ошибки, счетчик ошибок
3. Параллельная обработка
*/

func (c *Consumer) handleEvents(gotEvents []events.Event) error {
	var wg sync.WaitGroup

	eventsCh := make(chan events.Event, len(gotEvents))
	errorsCh := make(chan error, len(gotEvents))

	var wgErrors sync.WaitGroup
	wgErrors.Add(1)
	go func() {
		defer wgErrors.Done()
		for err := range errorsCh {
			log.Printf("[ERR] consumer: %s", err.Error())
		}
	}()

	for i := 0; i < c.workersCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("The panic has been intercepted: %v", r)
				}
			}()
			for event := range eventsCh {
				log.Printf("Processing event: %s", event.Text)
				if err := c.processor.Process(event); err != nil {
					errorsCh <- err
				}
			}
		}()
	}
	for _, event := range gotEvents {
		eventsCh <- event
	}

	close(eventsCh)

	wg.Wait()

	close(errorsCh)

	wgErrors.Wait()

	return nil
}
