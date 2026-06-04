# profiles

По результатам benchmark-ов потенциальным методом на оптимизацию стал RootHandler
- использование html/template
- есть вызов GetListGauges() и GetListCounters() = копирование map
- создание строк

## Алгоритм снятия профиля нагрузки
1. Запустить сервис
2. Заполнить метриками:
```bash
for i in {1..10000}
do
  curl -s -X POST http://localhost:8080/update/gauge/metric$i/$i > /dev/null
done

for i in {1..10000}
do
  curl -s -X POST http://localhost:8080/update/counters/metric$i/$i > /dev/null
done
```
3. Запустить hey на GET /
```bash
hey -n 1000 -c 50 http://localhost:8080/
```
4. Пока hey работает: снять профиль
```bash
curl   http://localhost:6060/debug/pprof/heap   -o profiles/base.pprof
```

## Основная проблема
Генерация HTML
Анализ профиля alloc_objects показал, что основное количество аллокаций создаётся при генерации HTML-страницы со списком метрик. Наибольший вклад вносят функции пакетов html/template и reflect, вызываемые при обработке большого количества элементов (~20000 строк таблицы). Аллокации в слое хранения метрик не являются основным источником потребления памяти.

## Решение
Шаблонизатор был заменён на генерацию HTML через strings.Builder, что позволило уменьшить количество создаваемых объектов примерно в 40 раз (с 243 млн до 5.8 млн по профилю alloc_objects) и исключить из профиля основные функции reflect.* и html/template.*.

## Итоги
BenchmarkRootHandler:
| Метрика   |         До |     После |
| --------- | ---------: | --------: |
| ns/op     | 15 654 593 | 3 049 918 |
| B/op      | 10 364 101 | 3 641 326 |
| allocs/op |    279 490 |    19 972 |

Профилирование памяти
| Тип профиля   |               До |            После |
| ------------- | ---------------: | ---------------: |
| alloc_objects | 243 млн объектов | 5.8 млн объектов |
| alloc_space   |           7.3 GB |          4.27 GB |

Type: alloc_objects
Time: 2026-06-03 01:56:42 MSK
Showing nodes accounting for -141056629, 96.02% of 146904212 total
Dropped 10 nodes (cum <= 734521)
      flat  flat%   sum%        cum   cum%
 -48890602 33.28% 33.28%  -48890602 33.28%  reflect.unsafe_New
 -22774455 15.50% 48.78%  -22774455 15.50%  html/template.htmlReplacer
 -22457516 15.29% 64.07% -138174529 94.06%  reflect.Value.call
 -22260904 15.15% 79.22%  -44019188 29.96%  reflect.MakeSlice
 -21758284 14.81% 94.04%  -21758284 14.81%  reflect.unsafe_NewArray
  -5242960  3.57% 97.60% -143419313 97.63%  text/template.(*state).walkRange
   1212435  0.83% 96.78%    1212435  0.83%  internal/strconv.FormatInt
   1114129  0.76% 96.02%    1114129  0.76%  internal/strconv.FormatFloat (inline)
       784 0.00053% 96.02%    1213219  0.83%  github.com/onbehalfofhim/metric-alert/internal/handler.mapToMetricView[go.shape.int64]
       744 0.00051% 96.02%    1114873  0.76%  github.com/onbehalfofhim/metric-alert/internal/handler.mapToMetricView[go.shape.float64]






