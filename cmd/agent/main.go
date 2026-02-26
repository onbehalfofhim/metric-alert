package main

import (
	"log"
	"time"

	"github.com/onbehalfofhim/metric-alert/internal/agent"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

func main() {
	// создание сборщика метрик
	collector := models.NewCollector()

	// создание клиента для отправки метрик
	client := agent.NewSender("http://localhost:8080")

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
			err := client.Send(metrics)
			if err != nil {
				log.Println("send failed:", err)
			}
		}
	}()
	// for _, v := range metrics {
	// 	fmt.Println(v)
	// }

	// чтобы main не завершился
	select {}
}
