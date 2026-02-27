package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	ulp "github.com/n0madic/go-ulp"
)

// writeTemplates outputs only the unique templates.
func writeTemplates(w io.Writer, result *ulp.ParseResult, format string) error {
	switch format {
	case "csv":
		return writeTemplatesCSV(w, result)
	case "json":
		return writeTemplatesJSON(w, result)
	case "text":
		return writeTemplatesText(w, result)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

// writeEvents outputs all events with their template assignments.
func writeEvents(w io.Writer, result *ulp.ParseResult, format string) error {
	switch format {
	case "csv":
		return writeEventsCSV(w, result)
	case "json":
		return writeEventsJSON(w, result)
	case "text":
		return writeEventsText(w, result)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

// CSV writers

func writeTemplatesCSV(w io.Writer, result *ulp.ParseResult) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	if err := cw.Write([]string{"Template", "Count"}); err != nil {
		return err
	}
	for _, t := range result.Templates {
		if err := cw.Write([]string{t.Template, strconv.Itoa(t.Count)}); err != nil {
			return err
		}
	}
	return cw.Error()
}

func writeEventsCSV(w io.Writer, result *ulp.ParseResult) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	if err := cw.Write([]string{"LineID", "EventID", "Content"}); err != nil {
		return err
	}
	for _, ev := range result.Events {
		if err := cw.Write([]string{
			strconv.Itoa(ev.LineID),
			ev.EventID,
			ev.RawContent,
		}); err != nil {
			return err
		}
	}
	return cw.Error()
}

// JSON writers

type templateJSON struct {
	Template string `json:"template"`
	Count    int    `json:"count"`
}

type eventJSON struct {
	LineID  int    `json:"line_id"`
	EventID string `json:"event_id"`
	Content string `json:"content"`
}

func writeTemplatesJSON(w io.Writer, result *ulp.ParseResult) error {
	items := make([]templateJSON, 0, len(result.Templates))
	for _, t := range result.Templates {
		items = append(items, templateJSON{
			Template: t.Template,
			Count:    t.Count,
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}

func writeEventsJSON(w io.Writer, result *ulp.ParseResult) error {
	items := make([]eventJSON, 0, len(result.Events))
	for _, ev := range result.Events {
		items = append(items, eventJSON{
			LineID:  ev.LineID,
			EventID: ev.EventID,
			Content: ev.RawContent,
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}

// Text writers

func writeTemplatesText(w io.Writer, result *ulp.ParseResult) error {
	for _, t := range result.Templates {
		if _, err := fmt.Fprintf(w, "(%d events) %s\n", t.Count, t.Template); err != nil {
			return err
		}
	}
	return nil
}

func writeEventsText(w io.Writer, result *ulp.ParseResult) error {
	for _, ev := range result.Events {
		if _, err := fmt.Fprintf(w, "%d\t%s\n", ev.LineID, ev.RawContent); err != nil {
			return err
		}
	}
	return nil
}
