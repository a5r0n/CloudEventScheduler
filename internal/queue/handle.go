package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/a5r0n/cloudeventscheduler/internal/task"
	"github.com/hibiken/asynq"
	"github.com/valyala/fasthttp"
)

var (
	httpClient = &fasthttp.Client{
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		// increase DNS cache time to an hour instead of default minute
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
	}
)

type Headers = map[string]string

func (q *Queue) handle(ctx context.Context, t *asynq.Task) error {
	var p task.WebhookTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	log.Printf(" [*] Send Webhook event to %s for %s %s", p.Url, t.Type(), t.ResultWriter().TaskID())
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(p.Url)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetBodyRaw([]byte(p.Body))

	var headers Headers
	if err := json.Unmarshal(p.Headers, &headers); err != nil {
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	req.Header.Set("X-Taks-Id", t.ResultWriter().TaskID())
	req.Header.Set("X-Task-TypeName", t.Type())
	req.Header.Set("X-Task-QueueName", "default")

	resp := fasthttp.AcquireResponse()

	err := httpClient.Do(req, resp)
	fasthttp.ReleaseRequest(req)

	if err == nil {
		result := strings.Replace(string(resp.Body()[:]), "", "", -1)

		log.Printf(" [*] Successfully sent Webhook event to %s: %s", p.Url, result)

		_, err := t.ResultWriter().Write([]byte(result))
		if err != nil {
			return fmt.Errorf("failed to write task result: %v", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "ERR Connection error: %v\n", err)
	}
	fasthttp.ReleaseResponse(resp)

	return nil
}
