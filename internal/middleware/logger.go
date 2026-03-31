package middleware

import (
	"net/http"
	"time"

	"github.com/onbehalfofhim/metric-alert/internal/logger"
)

type (
	// структура для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

// middleware для регистрации сведений о запросе
func RequestLogger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// функция Now() возвращает текущее время
			start := time.Now()

			responseData := &responseData{}
			lw := loggingResponseWriter{
				ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
				responseData:   responseData,
			}

			// точка, где выполняется хендлер
			h.ServeHTTP(&lw, r) // обслуживание оригинального запроса

			// Since возвращает разницу во времени между start
			// и моментом вызова Since. Таким образом можно посчитать
			// время выполнения запроса.
			duration := time.Since(start)

			// отправляем сведения о запросе в zap
			log.Info("request",
				"uri", r.RequestURI,
				"method", r.Method,
				"status", responseData.status,
				"duration", duration.String(),
				"size", responseData.size,
			)
		})
	}
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}
