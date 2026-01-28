package scanner

import (
	"database/sql"
	"errors"
	"github.com/controlplane-com/libs-go/pkg/pipeline"
)

type ScanResults [][]any

type DbScanner struct {
	MapperCollection *MapperCollection
}

func NewDbScanner() *DbScanner {
	return &DbScanner{
		MapperCollection: Mappers,
	}
}

func (s *DbScanner) ScanRowsIntoStructs(rows *sql.Rows, spacer string, models ...any) (ScanResults, error) {
	numModels := len(models)
	if models == nil || numModels == 0 {
		return nil, errors.New("models cannot be a nil or empty slice")
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var scanResults ScanResults
	for rows.Next() {
		scanTargets := s.buildDbScanTargets(models, columns, spacer)
		scanTargetFields := collateScanTargetFields(scanTargets)
		err = rows.Scan(scanTargetFields...)
		if err != nil {
			return nil, err
		}
		modelsInRow, err := pipeline.Map(scanTargets, func(st *ScanTarget[any]) (any, error) {
			return st.Model, nil
		})
		if err != nil {
			return nil, err
		}
		scanResults = append(scanResults, modelsInRow)
	}
	return scanResults, nil
}

func collateScanTargetFields(ts []*ScanTarget[any]) []any {
	fields := []any{}
	for _, t := range ts {
		for _, f := range t.fields {
			fields = append(fields, f)
		}
	}
	return fields
}

// Create a new list of scan targets using the given models as a template
func (s *DbScanner) buildDbScanTargets(models []any, columns []string, spacer string) []*ScanTarget[any] {
	numModels := len(models)
	modelIndex := -1
	scanTargets := make([]*ScanTarget[any], numModels)
	var t *ScanTarget[any]
	for _, c := range columns {
		if t == nil || s.detectSpacerColumn(c, spacer) {
			modelIndex++
			if modelIndex == numModels {
				break
			}
			t = CopyIntoNewScanTarget(models[modelIndex])
			scanTargets[modelIndex] = t
		}
		t.AddField(c, &AutoConversionMapper{})
	}
	return scanTargets
}

func (s *DbScanner) detectSpacerColumn(column string, spacer string) bool {
	lenColumn := len(column)
	lenSpacer := len(spacer)
	if lenColumn < lenSpacer {
		return false
	}
	return column[:lenSpacer] == spacer
}
