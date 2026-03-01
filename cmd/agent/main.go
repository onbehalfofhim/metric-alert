package main

import (
	"log"
	"strings"
	"time"

	"github.com/onbehalfofhim/metric-alert/internal/agent"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

func main() {
	parseFlags()

	// создание сборщика метрик
	collector := models.NewCollector()

	// создание клиента для отправки метрик
	if !strings.HasPrefix(flagRunAddr, "http://") {
		flagRunAddr = "http://" + flagRunAddr
	}

	client := agent.NewSender(flagRunAddr)

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
