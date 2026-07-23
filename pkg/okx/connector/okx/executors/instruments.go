package executors

import (
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/okx/connector/okx"
	"github.com/m1xar/scope360-reconstruction/pkg/okx/connector/okx/models"
)

const instrumentsPath = "/api/v5/public/instruments"

func FetchInstrument(client *resty.Client, baseURL, instID, instType string) (*models.Instrument, error) {
	params := map[string]string{
		"instType": instType,
		"instId":   instID,
	}

	data, err := okx.DoGet[[]models.Instrument](client, baseURL, instrumentsPath, params)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return &models.DefaultInstrument, nil
	}

	return &data[0], nil
}

func FetchInstruments(client *resty.Client, baseURL string, identifiers map[string]models.Instrumentidentifier) (map[string]models.Instrument, error) {
	var (
		wg          sync.WaitGroup
		mu          sync.Mutex
		err         error
		semaphore   = make(chan struct{}, 5)
		instruments = make(map[string]models.Instrument)
	)

	for i := range identifiers {
		if err != nil {
			break
		}

		wg.Add(1)
		id := identifiers[i]
		go func() {
			defer wg.Done()
			semaphore <- struct{}{}

			instrument, fetchErr := FetchInstrument(client, baseURL, id.InstID, id.InstType)
			if err != nil {
				err = fetchErr
				return
			}

			mu.Lock()
			instruments[id.InstID] = *instrument
			mu.Unlock()
		}()
		<-semaphore
	}
	wg.Wait()

	if err != nil {
		return instruments, err
	}

	return instruments, nil
}
