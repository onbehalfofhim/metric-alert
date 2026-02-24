package main

import (
	"time"

	"github.com/onbehalfofhim/metric-alert/internal/agent"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

func main() {
	// создание сборщика метрик
	collector := models.Collector{}
	collector.NewCollector()

	// создание клиента для отправки метрик
	client := agent.Sender{}
	client.NewSender("http://localhost:8080")

	// временные задержки для сборка метрик и отправки запроса
	var pollInterval int = 2
	var reportInterval int = 10

	// горутина для сборка метрик
	go func() {
		for {
			collector.CollectMetrics()
			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}()

	// горутина для отправки
	go func() {
		for {
			time.Sleep(time.Duration(reportInterval) * time.Second)

			metrics := collector.GetMetrics()
			client.Send(metrics)
		}
	}()
	// for _, v := range metrics {
	// 	fmt.Println(v)
	// }

	// чтобы main не завершился
	select {}
}
