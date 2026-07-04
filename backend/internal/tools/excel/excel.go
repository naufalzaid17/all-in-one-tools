// Package excel implements the Excel Toolkit tool: merging multiple
// .xlsx workbooks into one and converting sheets to CSV, backed by
// excelize.
//
// Workbooks are processed file-to-file inside the request workspace and
// rows are streamed (excelize row iterator + stream writer), so memory
// stays bounded even for large spreadsheets.
package excel

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/xuri/excelize/v2"

	"github.com/naufalzaid17/all-in-one-tools/backend/internal/core"
	"github.com/naufalzaid17/all-in-one-tools/backend/internal/fileio"
)

const (
	contentTypeXLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	contentTypeCSV  = "text/csv; charset=utf-8"
	// maxSheetName is Excel's hard limit for sheet names.
	maxSheetName = 31
)

var uploadOpts = fileio.ParseOptions{AllowedExtensions: []string{".xlsx"}}

// Tool is the Excel Toolkit tool module.
type Tool struct{}

// New returns the Excel Toolkit tool.
func New() *Tool { return &Tool{} }

func (t *Tool) ID() string              { return "excel" }
func (t *Tool) Name() string            { return "Excel Toolkit" }
func (t *Tool) Category() core.Category { return core.CategoryDocument }
func (t *Tool) Description() string {
	return "Merge .xlsx workbooks and convert sheets to CSV."
}

// RegisterRoutes mounts the tool endpoints under /api/v1/tools/excel.
func (t *Tool) RegisterRoutes(r chi.Router) {
	r.Post("/merge", t.handleMerge)
	r.Post("/to-csv", t.handleToCSV)
}

func parseUpload(w http.ResponseWriter, r *http.Request) (*fileio.Workspace, *fileio.Form, bool) {
	return fileio.ParseRequest(w, r, uploadOpts)
}

func (t *Tool) handleMerge(w http.ResponseWriter, r *http.Request) {
	ws, form, ok := parseUpload(w, r)
	if !ok {
		return
	}
	defer ws.Close()

	if len(form.Files) < 2 {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput, "merging requires at least two .xlsx files")
		return
	}

	outPath := ws.Path("merged.xlsx")
	if err := mergeWorkbooks(form.Files, outPath); err != nil {
		core.RespondError(w, http.StatusUnprocessableEntity, core.CodeInvalidInput, "merge failed: "+err.Error())
		return
	}
	core.RespondFileDownload(w, outPath, "merged.xlsx", contentTypeXLSX)
}

// mergeWorkbooks copies every sheet of every input workbook into a
// single output workbook. Cell values are carried over via streaming
// row iteration; formatting and formulas are intentionally flattened to
// values (documented behaviour of the merge tool).
func mergeWorkbooks(inputs []fileio.UploadedFile, outPath string) error {
	out := excelize.NewFile()
	defer out.Close()

	used := map[string]bool{}
	var firstSheet string

	for _, in := range inputs {
		src, err := excelize.OpenFile(in.Path)
		if err != nil {
			return fmt.Errorf("%s: not a valid .xlsx workbook", in.Name)
		}

		for _, sheet := range src.GetSheetList() {
			name := uniqueSheetName(used, in.Name, sheet)
			if _, err := out.NewSheet(name); err != nil {
				src.Close()
				return fmt.Errorf("create sheet %q: %w", name, err)
			}
			if firstSheet == "" {
				firstSheet = name
			}
			if err := copySheet(src, sheet, out, name); err != nil {
				src.Close()
				return fmt.Errorf("%s/%s: %w", in.Name, sheet, err)
			}
		}
		src.Close()
	}

	// Drop the implicit default sheet created by NewFile.
	if firstSheet != "" {
		if idx, err := out.GetSheetIndex(firstSheet); err == nil {
			out.SetActiveSheet(idx)
		}
		_ = out.DeleteSheet(out.GetSheetName(0))
	}
	return out.SaveAs(outPath)
}

func copySheet(src *excelize.File, srcSheet string, out *excelize.File, outSheet string) error {
	rows, err := src.Rows(srcSheet)
	if err != nil {
		return err
	}
	defer rows.Close()

	sw, err := out.NewStreamWriter(outSheet)
	if err != nil {
		return err
	}

	rowIdx := 0
	for rows.Next() {
		rowIdx++
		cols, err := rows.Columns()
		if err != nil {
			return err
		}
		if len(cols) == 0 {
			continue
		}
		cell, err := excelize.CoordinatesToCellName(1, rowIdx)
		if err != nil {
			return err
		}
		values := make([]any, len(cols))
		for i, c := range cols {
			values[i] = c
		}
		if err := sw.SetRow(cell, values); err != nil {
			return err
		}
	}
	return sw.Flush()
}

// uniqueSheetName builds an Excel-legal, workbook-unique sheet name
// that keeps the source file visible: "report.xlsx"+"Sheet1" →
// "report - Sheet1" (truncated to 31 chars, deduplicated with ~N).
func uniqueSheetName(used map[string]bool, fileName, sheetName string) string {
	base := strings.TrimSuffix(fileName, ".xlsx") + " - " + sheetName
	// Strip characters Excel forbids in sheet names.
	base = strings.Map(func(r rune) rune {
		switch r {
		case ':', '\\', '/', '?', '*', '[', ']':
			return ' '
		}
		return r
	}, base)
	if len(base) > maxSheetName {
		base = base[:maxSheetName]
	}
	name := base
	for i := 2; used[name]; i++ {
		suffix := fmt.Sprintf(" ~%d", i)
		name = base
		if len(name)+len(suffix) > maxSheetName {
			name = name[:maxSheetName-len(suffix)]
		}
		name += suffix
	}
	used[name] = true
	return name
}

func (t *Tool) handleToCSV(w http.ResponseWriter, r *http.Request) {
	ws, form, ok := parseUpload(w, r)
	if !ok {
		return
	}
	defer ws.Close()

	if len(form.Files) != 1 {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput, "exactly one .xlsx file is required")
		return
	}
	in := form.Files[0]

	src, err := excelize.OpenFile(in.Path)
	if err != nil {
		core.RespondError(w, http.StatusUnprocessableEntity, core.CodeInvalidInput, in.Name+" is not a valid .xlsx workbook")
		return
	}
	defer src.Close()

	sheet := form.Value("sheet")
	if sheet == "" {
		sheet = src.GetSheetName(0)
	}

	outFile, outPath, err := ws.CreateFile("converted.csv")
	if err != nil {
		core.RespondError(w, http.StatusInternalServerError, core.CodeInternal, "could not stage output file")
		return
	}
	err = writeSheetCSV(src, sheet, outFile)
	outFile.Close()
	if err != nil {
		core.RespondError(w, http.StatusUnprocessableEntity, core.CodeInvalidInput, err.Error())
		return
	}

	name := strings.TrimSuffix(in.Name, ".xlsx") + ".csv"
	core.RespondFileDownload(w, outPath, name, contentTypeCSV)
}

// writeSheetCSV streams one sheet's rows into CSV form.
func writeSheetCSV(src *excelize.File, sheet string, out io.Writer) error {
	rows, err := src.Rows(sheet)
	if err != nil {
		names := strings.Join(src.GetSheetList(), ", ")
		return fmt.Errorf("sheet %q not found (available: %s)", sheet, names)
	}
	defer rows.Close()

	cw := csv.NewWriter(out)
	for rows.Next() {
		record, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("read row: %w", err)
		}
		if record == nil {
			record = []string{}
		}
		if err := cw.Write(record); err != nil {
			return fmt.Errorf("write csv: %w", err)
		}
	}
	cw.Flush()
	return cw.Error()
}
