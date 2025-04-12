// SPDX-File-CopyrightText: (c) 2025 Andrei Ilin <ortfero@gmail.com>
// SPDX-License-Identifier: MIT

package orderstracker

import "time"

type ExecutionReportKind int

const (
	ReportNone ExecutionReportKind = iota
	ReportPlaced
	ReportModified
	ReportCanceled
	ReportFilled
	ReportRejected
)

type ExecutionReport struct {
	Kind    ExecutionReportKind
	Time    time.Time
	Message string
	Amount  uint64
	Price   uint64
}
