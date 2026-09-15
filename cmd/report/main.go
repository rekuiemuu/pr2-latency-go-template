package main

import (
	"encoding/csv"
	"fmt"
	"go-algo/telemetry"
	"log"
	"os"
	"strconv"
	"strings"
)

var names = []string{"baseline", "delay_50", "delay_100", "jitter", "loss_5", "combined"}
var colors = []string{"#2563eb", "#16a34a", "#ea580c", "#9333ea", "#dc2626", "#0891b2"}

func main() {
	f, err := os.Open("docs/latency_samples.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	reader := csv.NewReader(f)
	reader.Comma = ';'
	rows, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	data := map[string][]telemetry.Sample{}
	for _, r := range rows[1:] {
		number, _ := strconv.Atoi(r[1])
		seq, _ := strconv.Atoi(r[2])
		rtt, _ := strconv.ParseFloat(r[4], 64)
		srtt, _ := strconv.ParseFloat(r[5], 64)
		data[r[0]] = append(data[r[0]], telemetry.Sample{Series: r[0], Number: number, Sequence: uint16(seq), RTTMs: rtt, SRTTMs: srtt, Status: r[6]})
	}
	os.MkdirAll("docs/graphs", 0755)
	line(data)
	bars(data)
	histogram(data)
	report(data)
}

func begin(f *os.File, title string) {
	fmt.Fprintln(f, `<svg xmlns="http://www.w3.org/2000/svg" width="900" height="440" viewBox="0 0 900 440">`)
	fmt.Fprintln(f, `<rect width="900" height="440" fill="white"/><g font-family="Arial,sans-serif" font-size="13" fill="#222">`)
	fmt.Fprintf(f, `<text x="30" y="28" font-size="20">%s</text>`, title)
	fmt.Fprintln(f, `<path d="M65 45 V360 H850" fill="none" stroke="#555"/>`)
}
func end(f *os.File) { fmt.Fprintln(f, "</g></svg>"); f.Close() }

func line(data map[string][]telemetry.Sample) {
	f, err := os.Create("docs/graphs/rtt.svg")
	if err != nil {
		log.Fatal(err)
	}
	begin(f, "RTT по измерениям (красный крест = timeout)")
	for tick := 0; tick <= 200; tick += 50 {
		y := 360 - float64(tick)*1.45
		fmt.Fprintf(f, `<text x="23" y="%.1f">%d</text><path d="M65 %.1f H850" stroke="#ddd"/>`, y+4, tick, y)
	}
	fmt.Fprintln(f, `<text x="410" y="390">Номер PING</text><text x="5" y="220">мс</text>`)
	for i, name := range names {
		samples := data[name]
		path := strings.Builder{}
		for _, s := range samples {
			x := 65 + float64(s.Number-1)*15.8
			if s.Status != telemetry.Received {
				if s.Status == telemetry.TimedOut {
					fmt.Fprintf(f, `<path d="M%.1f 350 l8 8 m0 -8 l-8 8" stroke="red"/>`, x)
				}
				continue
			}
			y := 360 - s.RTTMs*1.45
			if path.Len() == 0 {
				fmt.Fprintf(&path, "M%.1f %.1f", x, y)
			} else {
				fmt.Fprintf(&path, " L%.1f %.1f", x, y)
			}
		}
		fmt.Fprintf(f, `<path d="%s" fill="none" stroke="%s" stroke-width="1.6"/><text x="%d" y="415" fill="%s">%s</text>`, path.String(), colors[i], 65+i*130, colors[i], name)
	}
	end(f)
}

func bars(data map[string][]telemetry.Sample) {
	f, err := os.Create("docs/graphs/summary.svg")
	if err != nil {
		log.Fatal(err)
	}
	begin(f, "Средний RTT, SRTT и потери")
	fmt.Fprintln(f, `<text x="15" y="48">мс</text><text x="850" y="48">%</text>`)
	for i, name := range names {
		s := telemetry.Calculate(data[name])
		x := 90 + i*125
		for j, v := range []float64{s.Mean, s.SRTT} {
			h := v * 2
			color := "#2563eb"
			if j == 1 {
				color = "#16a34a"
			}
			fmt.Fprintf(f, `<rect x="%d" y="%.1f" width="24" height="%.1f" fill="%s"/>`, x+j*27, 360-h, h, color)
		}
		h := s.LossRate * 40
		fmt.Fprintf(f, `<rect x="%d" y="%.1f" width="24" height="%.1f" fill="#dc2626"/><text x="%d" y="382">%s</text>`, x+54, 360-h, h, x-10, name)
	}
	fmt.Fprintln(f, `<text x="145" y="415" fill="#2563eb">■ mean RTT (мс)</text><text x="365" y="415" fill="#16a34a">■ SRTT (мс)</text><text x="540" y="415" fill="#dc2626">■ loss rate (%), правая шкала: 40 px/%</text>`)
	end(f)
}

func histogram(data map[string][]telemetry.Sample) {
	f, err := os.Create("docs/graphs/histogram.svg")
	if err != nil {
		log.Fatal(err)
	}
	begin(f, "Гистограмма RTT: baseline, jitter, combined")
	for i, name := range []string{"baseline", "jitter", "combined"} {
		bins := make([]int, 8)
		for _, s := range data[name] {
			if s.Status == telemetry.Received {
				bin := int(s.RTTMs) / 25
				if bin > 7 {
					bin = 7
				}
				bins[bin]++
			}
		}
		for j, n := range bins {
			h := float64(n) * 5.8
			x := 80 + j*93 + i*25
			fmt.Fprintf(f, `<rect x="%d" y="%.1f" width="22" height="%.1f" fill="%s"/>`, x, 360-h, h, colors[[]int{0, 3, 5}[i]])
		}
		fmt.Fprintf(f, `<text x="%d" y="415" fill="%s">%s</text>`, 250+i*180, colors[[]int{0, 3, 5}[i]], name)
	}
	for j := 0; j < 8; j++ {
		fmt.Fprintf(f, `<text x="%d" y="382">%d–%d</text>`, 75+j*93, j*25, (j+1)*25)
	}
	fmt.Fprintln(f, `<text x="650" y="400">RTT, мс (последний столбец ≥175)</text><text x="8" y="220">число</text>`)
	end(f)
}

func report(data map[string][]telemetry.Sample) {
	f, err := os.Create("docs/Latency_Report.md")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fmt.Fprintln(f, "# Отчёт о задержке UDP")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "Данные: [latency_samples.csv](latency_samples.csv), 2026-09-15, localhost, шесть реальных серий по 50 PING. Параметры указаны в [Experiment_Config.md](Experiment_Config.md).")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "| Серия | Отправлено | Получено | Timeout | min | max | mean | median | SRTT | джиттер | потери |")
	fmt.Fprintln(f, "|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|")
	for _, name := range names {
		s := telemetry.Calculate(data[name])
		fmt.Fprintf(f, "| %s | %d | %d | %d | %.3f | %.3f | %.3f | %.3f | %.3f | %.3f | %.1f%% |\n", name, s.Sent, s.Received, s.Timeouts, s.Min, s.Max, s.Mean, s.Median, s.SRTT, s.Jitter, s.LossRate)
	}
	fmt.Fprintln(f, "\nRTT, SRTT и джиттер — в миллисекундах. SRTT = 0.875·SRTT + 0.125·RTT (первый RTT инициализирует оценку). Джиттер — среднее |RTTᵢ − RTTᵢ₋₁| по успешным ответам; потери = timeout/sent·100%. Серверные часы в RTT не участвуют.")
	fmt.Fprintln(f, "\n![Линейный RTT; кресты обозначают тайм-ауты](graphs/rtt.svg)")
	fmt.Fprintln(f, "\n![Средний RTT, SRTT и потери](graphs/summary.svg)")
	fmt.Fprintln(f, "\n![Гистограмма baseline, jitter, combined](graphs/histogram.svg)")
	fmt.Fprintln(f, "\nПотери в loss_5 и combined составили 4%, а не ровно 5%: фиксированное правило каждый 20-й пакет на серии из 50 пакетов пропускает 2 ответа. Встроенная задержка явно указана в конфигурации; это измерение loopback, а не внешней сети.")
}
