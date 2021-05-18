/**
 * @license
 * Copyright (C) 2021  Radu Stefan
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see: https://www.gnu.org/licenses/
 **/

package report

import "github.com/360EntSecGroup-Skylar/excelize"

func getExcelStyle(name string, report *excelize.File) (style int) {
	switch name {
	case "subtitle":
		if style, err := report.NewStyle(`{
			"alignment":
			{
				"horizontal": "left",
				"vertical": "bottom"
			},
			"font":
			{
				"family": "Calibri body",
				"size": 14,
				"color": "#404040"
			}
		}`); err == nil {
			return style
		}
	case "currentValue":
		if style, err := report.NewStyle(`{
			"alignment":
			{
				"horizontal": "right",
				"vertical": "center"
			},
			"font":
			{
				"family": "Calibri body",
				"size": 32,
				"color": "#404040"
			},
			"number_format": 0
		}`); err == nil {
			return style
		}
	case "currentValueMoney":
		if style, err := report.NewStyle(`{
			"alignment":
			{
				"horizontal": "right",
				"vertical": "center"
			},
			"font":
			{
				"family": "Calibri body",
				"size": 32,
				"color": "#404040"
			},
			"number_format": 190,
			"decimal_places": 0
		}`); err == nil {
			return style
		}
	case "futureValue":
		if style, err := report.NewStyle(`{
			"alignment":
			{
				"horizontal": "center",
				"vertical": "center"
			},
			"font":
			{
				"family": "Calibri body",
				"size": 26,
				"color": "#404040"
			},
			"number_format": 0
		}`); err == nil {
			return style
		}
	case "futureValueMoney":
		if style, err := report.NewStyle(`{
			"alignment":
			{
				"horizontal": "center",
				"vertical": "center"
			},
			"font":
			{
				"family": "Calibri body",
				"size": 26,
				"color": "#404040"
			},
			"number_format": 190,
			"decimal_places": 0
		}`); err == nil {
			return style
		}
	case "subtitle2":
		if style, err := report.NewStyle(`{
			"alignment":
			{
				"horizontal": "center",
				"vertical": "bottom"
			},
			"font":
			{
				"family": "Calibri",
				"size": 14,
				"color": "#404040"
			}
		}`); err == nil {
			return style
		}
	case "valueExplain":
		if style, err := report.NewStyle(`{
			"alignment":
			{
				"horizontal": "left",
				"vertical": "center",
				"wrap_text": true
			},
			"font":
			{
				"family": "Calibri body",
				"size": 12,
				"color": "#404040"
			}
		}`); err == nil {
			return style
		}
	case "logoBump":
		if style, err := report.NewStyle(`{
			"alignment": {
				"horizontal": "left",
				"vertical": "bottom",
				"wrap_text": false
			},
			"font": {
				"family": "Bernina Sans Narrow Lt",
				"size": 40,
				"color": "#262626"
			}
		}`); err == nil {
			return style
		}
	case "summaryDetail":
		if style, err := report.NewStyle(`{
			"alignment": {
				"horizontal": "left",
				"vertical": "bottom",
				"wrap_text": false
			},
			"font": {
				"family": "Calibri",
				"size": 14,
				"color": "#404040"
			}
		}`); err == nil {
			return style
		}
	case "summaryMoney":
		if style, err := report.NewStyle(`{
			"alignment": {
				"horizontal": "center",
				"vertical": "bottom",
				"wrap_text": false
			},
			"font": {
				"family": "Calibri",
				"size": 12,
				"color": "#404040"
			},
			"number_format": 190,
			"decimal_places": 0
		}`); err == nil {
			return style
		}
	case "link":
		if style, err := report.NewStyle(`{
			"alignment":
			{
				"horizontal": "left",
				"vertical": "bottom",
				"wrap_text": false
			},
			"font":
			{
				"family": "Calibri",
				"size": 12,
				"color": "#0099CC",
				"underline": "single"
			}
		}`); err == nil {
			return style
		}
	default:
		if style, err := report.NewStyle(`{
			"alignment":
			{
				"horizontal": "left",
				"vertical": "bottom",
				"wrap_text": true
			},
			"font":
			{
				"family": "Calibri",
				"size": 12,
				"color": "#404040"
			}
		}`); err == nil {
			return style
		}
	}

	return -1
}
