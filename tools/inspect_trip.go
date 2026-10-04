package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

func main() {
	f, _ := os.Open("data/raw/combine_tije_opendata_cleaned.csv")
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = ';'
	r.Read()

	count := 0
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if rec[3] == "14A-R01" {
			fmt.Printf("Trip: %s, StopSeq: %4s, Stop: %-30s, Lat: %s, Lon: %s, ShapeSeq: %s, ShapeID: %s\n",
				rec[3], rec[6], rec[7], rec[8], rec[9], rec[11], rec[12])
			count++
		}
	}
}
