package svg

import (
	"fmt"
	"strings"
)

// GenerateMileageChartSVG generates an SVG area chart for mileage entries
func GenerateMileageChartSVG(entries []ChartPoint) string {
	if len(entries) < 2 {
		return `<div class="h-[180px] w-full flex items-center justify-center text-zinc-400">Not enough data for chart</div>`
	}

	// Find min/max for scaling
	var minOdometer, maxOdometer float64
	for i, e := range entries {
		if i == 0 || e.Odometer < minOdometer {
			minOdometer = e.Odometer
		}
		if i == 0 || e.Odometer > maxOdometer {
			maxOdometer = e.Odometer
		}
	}

	// Add padding
	padding := (maxOdometer - minOdometer) * 0.1
	if padding == 0 {
		padding = maxOdometer * 0.1
	}
	if padding == 0 {
		padding = 1000
	}
	minOdometer -= padding
	maxOdometer += padding

	// Chart dimensions (viewBox 0 0 100 100)
	width := 90.0
	height := 85.0
	marginLeft := 5.0
	marginBottom := 5.0

	// Generate points
	var points strings.Builder
	for i, e := range entries {
		x := marginLeft + (float64(i) / float64(len(entries)-1)) * width
		yRatio := (e.Odometer - minOdometer) / (maxOdometer - minOdometer)
		y := marginBottom + height - (yRatio * height)
		if i == 0 {
			fmt.Fprintf(&points, "%.1f,%.1f", x, y)
		} else {
			fmt.Fprintf(&points, " %.1f,%.1f", x, y)
		}
	}

	// Area points (add bottom corners for fill)
	areaPoints := points.String()
	if len(entries) > 0 {
		lastX := marginLeft + width
		areaPoints += fmt.Sprintf(" %.1f,%.1f %.1f,%.1f", lastX, marginBottom+height, marginLeft, marginBottom+height)
	}

	// Grid lines (horizontal)
	var gridLines strings.Builder
	for i := 0; i <= 5; i++ {
		y := marginBottom + height - (float64(i)/5.0)*height
		fmt.Fprintf(&gridLines, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="rgba(200,200,200,0.15)" stroke-dasharray="3,3" stroke-width="0.5"/>`,
			marginLeft, y, marginLeft+width, y)
	}

	return fmt.Sprintf(`<svg class="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
    <defs>
        <linearGradient id="mileageGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%%" stop-color="#4f46e5" stop-opacity="0.2" />
            <stop offset="95%%" stop-color="#4f46e5" stop-opacity="0" />
        </linearGradient>
    </defs>
    %s
    <polyline fill="none" stroke="#4f46e5" stroke-width="2" points="%s" />
    <polyline fill="url(#mileageGradient)" stroke="none" points="%s" />
</svg>`, gridLines.String(), points.String(), areaPoints)
}

type ChartPoint struct {
	Date      string
	Odometer  float64
	Expense   float64
}

// FormatChartPoints generates SVG points string from mileage entries
func FormatChartPoints(entries []ChartPoint) string {
	if len(entries) == 0 {
		return ""
	}

	// Find min/max
	var minOdometer, maxOdometer float64
	for i, e := range entries {
		if i == 0 || e.Odometer < minOdometer {
			minOdometer = e.Odometer
		}
		if i == 0 || e.Odometer > maxOdometer {
			maxOdometer = e.Odometer
		}
	}

	padding := (maxOdometer - minOdometer) * 0.1
	if padding == 0 {
		padding = maxOdometer * 0.1
	}
	if padding == 0 {
		padding = 1000
	}
	minOdometer -= padding
	maxOdometer += padding

	width := 90.0
	height := 85.0
	marginLeft := 5.0
	marginBottom := 5.0

	var sb strings.Builder
	for i, e := range entries {
		x := marginLeft + (float64(i)/float64(len(entries)-1))*width
		yRatio := (e.Odometer - minOdometer) / (maxOdometer - minOdometer)
		y := marginBottom + height - (yRatio * height)
		if i == 0 {
			fmt.Fprintf(&sb, "%.1f,%.1f", x, y)
		} else {
			fmt.Fprintf(&sb, " %.1f,%.1f", x, y)
		}
	}
	return sb.String()
}