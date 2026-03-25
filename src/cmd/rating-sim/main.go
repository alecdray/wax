package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"golang.org/x/term"
)

type params struct {
	floor    float64
	ceiling  float64
	exponent float64
}

func curvedValue(value int, exponent float64) float64 {
	normalized := (float64(value) - 3.0) / 2.0
	curved := math.Copysign(math.Pow(math.Abs(normalized), exponent), normalized)
	return curved*2.0 + 3.0
}

func computeRating(consistency, impact, gutCheck int, p params) float64 {
	raw := curvedValue(consistency, p.exponent)*0.3 +
		curvedValue(impact, p.exponent)*0.4 +
		curvedValue(gutCheck, p.exponent)*0.3
	rating := p.floor + ((raw-1.0)/4.0)*(p.ceiling-p.floor)
	return math.Round(rating*10) / 10
}

type labelEntry struct {
	min   float64
	max   float64
	label string
}

var ratingKey = []labelEntry{
	{0.0, 2.9, "DOA"},
	{3.0, 3.9, "Nope"},
	{4.0, 4.9, "Not For Me"},
	{5.0, 5.9, "Lukewarm"},
	{6.0, 6.9, "Solid"},
	{7.0, 7.9, "Staff Pick"},
	{8.0, 8.9, "Essential"},
	{9.0, 9.9, "Inst. Classic"},
	{10.0, 10.0, "Masterpiece"},
}

func getLabel(rating float64) string {
	clamped := math.Max(0, math.Min(10, rating))
	for _, e := range ratingKey {
		if clamped >= e.min && clamped <= e.max {
			return e.label
		}
	}
	return ratingKey[len(ratingKey)-1].label
}

func distribution(p params) [9]int {
	var counts [9]int
	for c := 1; c <= 5; c++ {
		for i := 1; i <= 5; i++ {
			for g := 1; g <= 5; g++ {
				r := computeRating(c, i, g, p)
				for idx, e := range ratingKey {
					if r >= e.min && r <= e.max {
						counts[idx]++
						break
					}
				}
			}
		}
	}
	return counts
}

func bar(count, max int, width int) string {
	if max == 0 {
		return strings.Repeat("░", width)
	}
	filled := int(math.Round(float64(count) / float64(max) * float64(width)))
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func clamp(v, lo, hi float64) float64 {
	return math.Max(lo, math.Min(hi, v))
}

func render(p params) string {
	var b strings.Builder

	// header
	b.WriteString("\033[2J\033[H") // clear + home
	b.WriteString("  Rating Score Simulator\n")
	b.WriteString("  " + strings.Repeat("─", 52) + "\n\n")

	// parameters
	b.WriteString("  Parameters\n")
	b.WriteString(fmt.Sprintf("    Floor:    %4.1f    [f / F  ±0.5]\n", p.floor))
	b.WriteString(fmt.Sprintf("    Ceiling: %5.1f    [c / C  ±0.5]\n", p.ceiling))
	b.WriteString(fmt.Sprintf("    Exponent: %4.2f    [e / E  ±0.05]\n", p.exponent))
	b.WriteString("\n")

	// curved values
	b.WriteString("  Curved values per answer\n")
	b.WriteString("    ")
	for v := 1; v <= 5; v++ {
		cv := curvedValue(v, p.exponent)
		b.WriteString(fmt.Sprintf("%d → %4.2f   ", v, cv))
	}
	b.WriteString("\n\n")

	// uniform benchmarks
	b.WriteString("  Uniform answer scores\n")
	for v := 1; v <= 5; v++ {
		score := computeRating(v, v, v, p)
		label := getLabel(score)
		b.WriteString(fmt.Sprintf("    All %ds  → %5.1f  %s\n", v, score, label))
	}
	b.WriteString("\n")

	// label distribution
	counts := distribution(p)
	maxCount := 0
	for _, c := range counts {
		if c > maxCount {
			maxCount = c
		}
	}

	b.WriteString("  Label distribution (125 combos)\n")
	for i, e := range ratingKey {
		barchart := bar(counts[i], maxCount, 20)
		b.WriteString(fmt.Sprintf("    %-14s  %s  %3d\n", e.label, barchart, counts[i]))
	}

	b.WriteString("\n  [q] quit\n")
	return b.String()
}

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to set raw mode:", err)
		os.Exit(1)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	defer fmt.Print("\033[?25h") // restore cursor

	fmt.Print("\033[?25l") // hide cursor

	p := params{floor: 2.0, ceiling: 10.0, exponent: 0.8}
	fmt.Print(render(p))

	buf := make([]byte, 3)
	for {
		n, _ := os.Stdin.Read(buf)
		if n == 0 {
			continue
		}
		switch buf[0] {
		case 'q', 'Q', 3: // q, Q, ctrl-c
			fmt.Print("\033[2J\033[H")
			return
		case 'f':
			p.floor = clamp(p.floor-0.5, 0, p.ceiling-1)
		case 'F':
			p.floor = clamp(p.floor+0.5, 0, p.ceiling-1)
		case 'c':
			p.ceiling = clamp(p.ceiling-0.5, p.floor+1, 20)
		case 'C':
			p.ceiling = clamp(p.ceiling+0.5, p.floor+1, 20)
		case 'e':
			p.exponent = clamp(math.Round((p.exponent-0.05)*100)/100, 0.05, 2.0)
		case 'E':
			p.exponent = clamp(math.Round((p.exponent+0.05)*100)/100, 0.05, 2.0)
		}
		fmt.Print(render(p))
	}
}
