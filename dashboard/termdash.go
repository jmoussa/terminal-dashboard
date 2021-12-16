package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/sparkline"
	"github.com/pbnjay/memory"
	"github.com/shirou/gopsutil/cpu"
)

// playType indicates how to play a donut.
type playType int

const (
	playTypePercent playType = iota
	playTypeAbsolute
)

func getCPU() int {
	percentage, err := cpu.Percent(0, true)
	if err != nil {
		panic(err)
	}
	metric := 0
	for idx, cpupercent := range percentage {
		if strconv.Itoa(idx) == "0" {
			metric = int(cpupercent)
		}
	}
	return metric
}

func getMemory() uint64 {
	var system_memory uint64 = memory.TotalMemory()
	dt := time.Now().String()
	fmt.Println(dt + ": " + strconv.Itoa(int(system_memory)))
	return system_memory
}

// playDonut continuously changes the displayed percent value on the donut by the
// step once every delay. Exits when the context expires.
func playDonut(ctx context.Context, d *donut.Donut, start, step int, delay time.Duration, pt playType) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			switch pt {
			case playTypePercent:
				metric := getCPU()
				if err := d.Percent(metric); err != nil {
					panic(err)
				}
			case playTypeAbsolute:
				metric := getCPU()
				if err := d.Absolute(metric, 100); err != nil {
					panic(err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

// playSparkLine continuously adds values to the SparkLine, once every delay.
// Exits when the context expires.
func playSparkLine(ctx context.Context, sl *sparkline.SparkLine, delay time.Duration) {

	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			v := int(getCPU())
			if err := sl.Add([]int{v}); err != nil {
				panic(err)
			}

		case <-ctx.Done():
			return
		}
	}
}

func main() {
	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()
	const redrawInterval = 250 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())

	// create new donut for cpu usage
	local_cpu, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorGreen)),
		donut.Label("CPU Percentage - localhost", cell.FgColor(cell.ColorGreen)),
	)
	if err != nil {
		panic(err)
	}
	// create new sparkline for local cpu usage
	local_cpu_sparkline, err := sparkline.New(
		sparkline.Label("CPU Sparkline - localhost", cell.FgColor(cell.ColorNumber(33))),
		sparkline.Color(cell.ColorGreen),
	)
	if err != nil {
		panic(err)
	}

	// Start concurrent donut and sparkline data flow
	go playDonut(ctx, local_cpu, 0, 1, redrawInterval/3, playTypePercent)
	go playSparkLine(ctx, local_cpu_sparkline, redrawInterval/3)

	// Set up visual UI display
	terminalContainer, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.PlaceWidget(local_cpu),
						container.MarginLeft(2),
						container.MarginRight(2),
						container.MarginTop(2),
						container.Border(linestyle.Light),
					),
					container.Bottom(container.PlaceWidget(local_cpu_sparkline),
						container.MarginLeft(3),
						container.MarginRight(3),
						container.MarginBottom(5),
						container.MarginTop(5),
						container.Border(linestyle.Light),
					),
				),
			),
			container.Right(
				container.SplitVertical(
					container.Left(container.PlaceWidget(local_cpu)),
					container.Right(container.PlaceWidget(local_cpu_sparkline)),
				),
			),
		),
	)
	if err != nil {
		panic(err)
	}

	// Set exit key ('Q' || 'q')
	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	// Run terminal dashboard
	if err := termdash.Run(ctx, t, terminalContainer, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		panic(err)
	}
}
