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

import (
	"embed"
	"fmt"
	_ "image/png"
	"os"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/config"
	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/environment"
	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/util"
	"github.com/spf13/afero"
)

type useCaseData struct {
	mainValueMeaning       string
	mainValueDescription   string
	secondValueMeaning     string
	secondValueDescription string
	mainIcon               string
	secondIcon             string
}

//go:embed img
var imgs embed.FS
var allUseCases = make(map[string]useCaseData)

func CreateReport(env environment.Environment, config config.Config, reportData map[string]map[string]interface{}, outputDir string, fs afero.Fs) error {
	reportDir := outputDir + string(os.PathSeparator) + env.GetName()
	reportPath := reportDir + string(os.PathSeparator) + util.GetTodayDigitString() + "_" + config.GetId() + ".xlsx"

	if exists, err := afero.DirExists(fs, outputDir); !exists && err == nil {
		if err := fs.MkdirAll(outputDir, 0777); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if exists, err := afero.DirExists(fs, reportDir); !exists && err == nil {
		if err := fs.Mkdir(reportDir, 0777); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	report := excelize.NewFile()

	report.SetSheetName("Sheet1", "Summary")
	populateSummarySheet("Summary", report, reportData, env, config)
	report.SetSheetViewOptions("Summary", 0, excelize.ShowGridLines(false))

	for envErr, details := range reportData {
		idx := report.NewSheet(envErr)
		report.SetActiveSheet(idx)

		// layout
		report.SetSheetViewOptions(envErr, 0, excelize.ShowGridLines(false))
		setColumnWidths(envErr, report)
		populateBaseData(envErr, report)

		// values that are always present
		report.SetCellValue(envErr, "D2", envErr)
		report.SetCellValue(envErr, "C6", details["impacted_users"].(int))
		report.SetCellValue(envErr, "G6", details["unconverted_users"].(int))
		report.SetCellValue(envErr, "K6", details["lost_users"].(int))

		// charts
		addUserBreakdownChart(envErr, report, details["user_breakdown"], "B12")
		addDailyBreakdownChart(envErr, report, details["date_breakdown"], "F12")

		// values that are populated based on use case configuration
		populateUseCaseCurrentData(envErr, report, config, details)
		populateUseCaseFutureData(envErr, report, config, details)
	}

	report.SetActiveSheet(0)

	if err := report.SaveAs(reportPath); err != nil {
		return err
	}

	return nil
}

func setColumnWidths(sheet string, report *excelize.File) {
	// Separator columns
	report.SetColWidth(sheet, "A", "A", 2.43)
	report.SetColWidth(sheet, "E", "E", 5.57)
	report.SetColWidth(sheet, "I", "I", 5.57)
	// Icon columns
	report.SetColWidth(sheet, "B", "B", 16)
	report.SetColWidth(sheet, "F", "F", 16)
	report.SetColWidth(sheet, "J", "J", 16)
	// Results columns
	report.SetColWidth(sheet, "C", "C", 24.57)
	report.SetColWidth(sheet, "G", "G", 24.57)
	report.SetColWidth(sheet, "K", "K", 24.57)
	// Medium text columns
	report.SetColWidth(sheet, "D", "D", 14.43)
	report.SetColWidth(sheet, "H", "H", 14.43)
	report.SetColWidth(sheet, "L", "L", 14.43)
}

func populateBaseData(sheet string, report *excelize.File) {
	// Style
	styleCurrentValue := getExcelStyle("currentValue", report)
	styleSubtitle := getExcelStyle("subtitle", report)
	styleValueExplain := getExcelStyle("valueExplain", report)
	styleDefault := getExcelStyle("default", report)
	report.MergeCell(sheet, "B6", "B10")
	report.MergeCell(sheet, "C6", "C7")
	report.MergeCell(sheet, "C9", "D10")
	report.MergeCell(sheet, "D2", "I2")
	report.MergeCell(sheet, "D6", "D7")
	report.MergeCell(sheet, "F6", "F10")
	report.MergeCell(sheet, "G6", "G7")
	report.MergeCell(sheet, "G9", "H10")
	report.MergeCell(sheet, "H6", "H7")
	report.MergeCell(sheet, "J6", "J10")
	report.MergeCell(sheet, "K6", "K7")
	report.MergeCell(sheet, "K9", "L10")
	report.MergeCell(sheet, "L6", "L7")
	report.SetCellStyle(sheet, "B2", "B2", styleSubtitle)
	report.SetCellStyle(sheet, "D2", "D2", styleValueExplain)
	report.SetCellStyle(sheet, "B4", "B4", styleSubtitle)
	report.SetCellStyle(sheet, "C6", "C6", styleCurrentValue)
	report.SetCellStyle(sheet, "D6", "D6", styleValueExplain)
	report.SetCellStyle(sheet, "C9", "C9", styleDefault)
	report.SetCellStyle(sheet, "G6", "G6", styleCurrentValue)
	report.SetCellStyle(sheet, "H6", "H6", styleValueExplain)
	report.SetCellStyle(sheet, "G9", "G9", styleDefault)
	report.SetCellStyle(sheet, "K6", "K6", styleCurrentValue)
	report.SetCellStyle(sheet, "L6", "L6", styleValueExplain)
	report.SetCellStyle(sheet, "K9", "K9", styleDefault)

	// Text
	report.SetCellValue(sheet, "B2", "Error Analysed")
	report.SetCellValue(sheet, "B4", "Based on the last 7 days...")
	report.SetCellValue(sheet, "D6", "Impacted users")
	report.SetCellValue(sheet, "C9", "Total number of users that received the error.")
	report.SetCellValue(sheet, "H6", "Unconverted users")
	report.SetCellValue(sheet, "G9", "Of the impacted users, the number who did not convert in that session.")
	report.SetCellValue(sheet, "L6", "Lost users")
	report.SetCellValue(sheet, "K9", "Of the users who did not convert, how many did not return later to convert.")
	// Icons
	if icBlueUser, err := imgs.ReadFile("img/user_blue.png"); err == nil {
		report.AddPictureFromBytes(sheet, "B6", `{
			"lock_aspect_ratio": true,
			"positioning": "oneCell",
			"x_scale": 0.23,
			"y_scale": 0.23,
			"y_offset": -8
		}`, "user_blue", ".png", icBlueUser)
	}
	if icPurpleUser, err := imgs.ReadFile("img/user_purple.png"); err == nil {
		report.AddPictureFromBytes(sheet, "F6", `{
			"lock_aspect_ratio": true,
			"positioning": "oneCell",
			"x_scale": 0.23,
			"y_scale": 0.23,
			"y_offset": -8
		}`, "user_purple", ".png", icPurpleUser)
	}
	if icGreenUser, err := imgs.ReadFile("img/user_green.png"); err == nil {
		report.AddPictureFromBytes(sheet, "J6", `{
			"lock_aspect_ratio": true,
			"positioning": "oneCell",
			"x_scale": 0.23,
			"y_scale": 0.23,
			"y_offset": -8
		}`, "user_green", ".png", icGreenUser)
	}
}

func populateUseCaseCurrentData(sheet string, report *excelize.File, config config.Config, details map[string]interface{}) {
	styleCurrentValue := getExcelStyle("currentValue", report)
	styleCurrentValueMoney := getExcelStyle("currentValueMoney", report)
	styleSubtitle := getExcelStyle("subtitle", report)
	styleValueExplain := getExcelStyle("valueExplain", report)
	styleDefault := getExcelStyle("default", report)

	// Define flat text that accompanies use case data in the report
	allUseCases["lost_basket"] = useCaseData{
		mainValueMeaning:       "Revenue at risk",
		mainValueDescription:   "Total basket value potentially at risk from the lost users.",
		secondValueMeaning:     "Potential lost profit",
		secondValueDescription: "The total profit potentially lost based on a {MARGIN}% margin.",
		mainIcon:               "cart_purple.png",
		secondIcon:             "money_purple.png",
	}
	allUseCases["agent_hours"] = useCaseData{
		mainValueMeaning:       "Agent hours lost",
		mainValueDescription:   "Estimated number of hours spent by agents on calls as a result of the error.",
		secondValueMeaning:     "Potential cost of calls",
		secondValueDescription: "The estimated cost of agents handling calls associated with the error.",
		mainIcon:               "time_blue.png",
		secondIcon:             "money_blue.png",
	}
	allUseCases["incurred_costs"] = useCaseData{
		mainValueMeaning:     "Potential costs incurred",
		mainValueDescription: "The estimated revenue impact from users hitting the error in your application.",
		mainIcon:             "money_green.png",
	}

	report.SetCellValue(sheet, "B24", "Based on the "+fmt.Sprintf("%d", details["lost_users"].(int))+" lost users...")
	report.SetCellStyle(sheet, "B24", "B24", styleSubtitle)

	// Add the text to the report depending on the use cases analysed
	for i, useCase := range config.GetUseCases() {
		idx := 26 + 6*i

		// Style
		report.MergeCell(sheet, "B"+fmt.Sprintf("%d", idx), "B"+fmt.Sprintf("%d", idx+4))
		report.MergeCell(sheet, "C"+fmt.Sprintf("%d", idx), "C"+fmt.Sprintf("%d", idx+1))
		report.MergeCell(sheet, "C"+fmt.Sprintf("%d", idx+3), "D"+fmt.Sprintf("%d", idx+4))
		report.MergeCell(sheet, "D"+fmt.Sprintf("%d", idx), "D"+fmt.Sprintf("%d", idx+1))
		report.MergeCell(sheet, "F"+fmt.Sprintf("%d", idx), "F"+fmt.Sprintf("%d", idx+4))
		report.MergeCell(sheet, "G"+fmt.Sprintf("%d", idx), "G"+fmt.Sprintf("%d", idx+1))
		report.MergeCell(sheet, "G"+fmt.Sprintf("%d", idx+3), "H"+fmt.Sprintf("%d", idx+4))
		report.MergeCell(sheet, "H"+fmt.Sprintf("%d", idx), "H"+fmt.Sprintf("%d", idx+1))
		report.SetCellStyle(sheet, "C"+fmt.Sprintf("%d", idx), "C"+fmt.Sprintf("%d", idx), styleCurrentValueMoney)
		report.SetCellStyle(sheet, "D"+fmt.Sprintf("%d", idx), "D"+fmt.Sprintf("%d", idx), styleValueExplain)
		report.SetCellStyle(sheet, "C"+fmt.Sprintf("%d", idx+3), "C"+fmt.Sprintf("%d", idx+3), styleDefault)
		report.SetCellStyle(sheet, "G"+fmt.Sprintf("%d", idx), "G"+fmt.Sprintf("%d", idx), styleCurrentValueMoney)
		report.SetCellStyle(sheet, "H"+fmt.Sprintf("%d", idx), "H"+fmt.Sprintf("%d", idx), styleValueExplain)
		report.SetCellStyle(sheet, "G"+fmt.Sprintf("%d", idx+3), "G"+fmt.Sprintf("%d", idx+3), styleDefault)

		// Flat texts and icons
		// Main results
		report.SetCellValue(sheet, "D"+fmt.Sprintf("%d", idx), allUseCases[string(useCase)].mainValueMeaning)
		report.SetCellValue(sheet, "C"+fmt.Sprintf("%d", idx+3), allUseCases[string(useCase)].mainValueDescription)
		if mainIcon, err := imgs.ReadFile("img/" + allUseCases[string(useCase)].mainIcon); err == nil {
			report.AddPictureFromBytes(sheet, "B"+fmt.Sprintf("%d", idx), `{
				"lock_aspect_ratio": true,
				"positioning": "oneCell",
				"x_scale": 0.23,
				"y_scale": 0.23,
				"y_offset": -8
			}`, allUseCases[string(useCase)].mainIcon, ".png", mainIcon)
		}
		// Secondary results, if any
		if allUseCases[string(useCase)].secondValueMeaning != "" {
			report.SetCellValue(sheet, "H"+fmt.Sprintf("%d", idx), allUseCases[string(useCase)].secondValueMeaning)
			if secondIcon, err := imgs.ReadFile("img/" + allUseCases[string(useCase)].secondIcon); err == nil {
				report.AddPictureFromBytes(sheet, "F"+fmt.Sprintf("%d", idx), `{
					"lock_aspect_ratio": true,
					"positioning": "oneCell",
					"x_scale": 0.23,
					"y_scale": 0.23,
					"y_offset": -8
				}`, allUseCases[string(useCase)].secondIcon, ".png", secondIcon)
			}
			if useCase == "lost_basket" {
				if margin := config.GetProperty("margin"); margin != nil {
					report.SetCellValue(sheet, "G"+fmt.Sprintf("%d", idx+3), strings.Replace(allUseCases["lost_basket"].secondValueDescription, "{MARGIN}", fmt.Sprintf("%.1f", margin.(float64)), 1))
				} else {
					report.SetCellValue(sheet, "G"+fmt.Sprintf("%d", idx+3), strings.Replace(allUseCases["lost_basket"].secondValueDescription, "{MARGIN}", "15", 1))
				}
			} else {
				report.SetCellValue(sheet, "G"+fmt.Sprintf("%d", idx+3), allUseCases[string(useCase)].secondValueDescription)
			}
		}
		// Values
		var mainValue interface{}
		var secondValue interface{}
		if useCase == "lost_basket" {
			mainValue = fmt.Sprintf("£%d", details["lost_basket"].(int))
			secondValue = fmt.Sprintf("£%d", details["lost_money"].(int))
		} else if useCase == "agent_hours" {
			report.SetCellStyle(sheet, "C"+fmt.Sprintf("%d", idx), "C"+fmt.Sprintf("%d", idx), styleCurrentValue)
			mainValue = details["lost_agent_hours"].(float64)
			if val, ok := details["hours_lost_cost"]; ok {
				secondValue = fmt.Sprintf("£%.0f", val.(float64))
			}
		} else if useCase == "incurred_costs" {
			mainValue = fmt.Sprintf("£%d", details["costs_incurred"].(int))
		}
		report.SetCellValue(sheet, "C"+fmt.Sprintf("%d", idx), mainValue)
		if allUseCases[string(useCase)].secondValueMeaning != "" && secondValue != "" {
			report.SetCellValue(sheet, "G"+fmt.Sprintf("%d", idx), secondValue)
		}

		// Separators
		report.SetRowHeight(sheet, idx+2, 7) // Small separator row
		if i != 0 {
			report.SetRowHeight(sheet, idx-1, 22) // Bigger separator row
		}
	}
}

func populateUseCaseFutureData(sheet string, report *excelize.File, config config.Config, details map[string]interface{}) {
	styleSubtitle := getExcelStyle("subtitle", report)
	styleSubtitle2 := getExcelStyle("subtitle2", report)
	styleCurrentValue := getExcelStyle("currentValue", report)
	styleFutureValueMoney := getExcelStyle("futureValueMoney", report)
	styleDefault := getExcelStyle("default", report)

	offset := len(config.GetUseCases()) * 6
	idx := 26 + offset

	// Flat text and styles
	report.SetCellStyle(sheet, "B"+fmt.Sprintf("%d", idx), "B"+fmt.Sprintf("%d", idx), styleSubtitle)
	report.SetCellValue(sheet, "B"+fmt.Sprintf("%d", idx), "Potential business impact over the next...")
	idx += 2
	report.MergeCell(sheet, "B"+fmt.Sprintf("%d", idx), "B"+fmt.Sprintf("%d", idx+1))
	report.MergeCell(sheet, "F"+fmt.Sprintf("%d", idx), "F"+fmt.Sprintf("%d", idx+1))
	report.MergeCell(sheet, "J"+fmt.Sprintf("%d", idx), "J"+fmt.Sprintf("%d", idx+1))
	report.SetCellStyle(sheet, "B"+fmt.Sprintf("%d", idx), "B"+fmt.Sprintf("%d", idx), styleCurrentValue)
	report.SetCellStyle(sheet, "C"+fmt.Sprintf("%d", idx), "C"+fmt.Sprintf("%d", idx), styleDefault)
	report.SetCellStyle(sheet, "F"+fmt.Sprintf("%d", idx), "F"+fmt.Sprintf("%d", idx), styleCurrentValue)
	report.SetCellStyle(sheet, "G"+fmt.Sprintf("%d", idx), "G"+fmt.Sprintf("%d", idx), styleDefault)
	report.SetCellStyle(sheet, "J"+fmt.Sprintf("%d", idx), "J"+fmt.Sprintf("%d", idx), styleCurrentValue)
	report.SetCellStyle(sheet, "K"+fmt.Sprintf("%d", idx), "K"+fmt.Sprintf("%d", idx), styleDefault)
	report.SetCellValue(sheet, "B"+fmt.Sprintf("%d", idx), 14)
	report.SetCellValue(sheet, "C"+fmt.Sprintf("%d", idx), "days")
	report.SetCellValue(sheet, "F"+fmt.Sprintf("%d", idx), 21)
	report.SetCellValue(sheet, "G"+fmt.Sprintf("%d", idx), "days")
	report.SetCellValue(sheet, "J"+fmt.Sprintf("%d", idx), 28)
	report.SetCellValue(sheet, "K"+fmt.Sprintf("%d", idx), "days")
	idx += 3

	// Add the text to the report depending on the use cases analysed
	for _, useCase := range config.GetUseCases() {
		report.SetRowHeight(sheet, idx, 18)
		report.SetRowHeight(sheet, idx+1, 33.75)
		report.MergeCell(sheet, "B"+fmt.Sprintf("%d", idx), "C"+fmt.Sprintf("%d", idx))
		report.MergeCell(sheet, "B"+fmt.Sprintf("%d", idx+1), "C"+fmt.Sprintf("%d", idx+1))
		report.MergeCell(sheet, "F"+fmt.Sprintf("%d", idx+1), "G"+fmt.Sprintf("%d", idx+1))
		report.MergeCell(sheet, "J"+fmt.Sprintf("%d", idx+1), "K"+fmt.Sprintf("%d", idx+1))
		report.SetCellStyle(sheet, "B"+fmt.Sprintf("%d", idx), "B"+fmt.Sprintf("%d", idx), styleSubtitle2)
		report.SetCellStyle(sheet, "B"+fmt.Sprintf("%d", idx+1), "B"+fmt.Sprintf("%d", idx+1), styleFutureValueMoney)
		report.SetCellStyle(sheet, "F"+fmt.Sprintf("%d", idx+1), "F"+fmt.Sprintf("%d", idx+1), styleFutureValueMoney)
		report.SetCellStyle(sheet, "J"+fmt.Sprintf("%d", idx+1), "J"+fmt.Sprintf("%d", idx+1), styleFutureValueMoney)

		if useCase == "lost_basket" {
			report.SetCellValue(sheet, "B"+fmt.Sprintf("%d", idx), "Due to lost baskets...")
			report.SetCellValue(sheet, "B"+fmt.Sprintf("%d", idx+1), fmt.Sprintf("£%d", details["lost_money_14d"].(int)))
			report.SetCellValue(sheet, "F"+fmt.Sprintf("%d", idx+1), fmt.Sprintf("£%d", details["lost_money_21d"].(int)))
			report.SetCellValue(sheet, "J"+fmt.Sprintf("%d", idx+1), fmt.Sprintf("£%d", details["lost_money_28d"].(int)))
			idx += 3
		} else if useCase == "agent_hours" {
			// Additional cells for this use case
			report.MergeCell(sheet, "B"+fmt.Sprintf("%d", idx+2), "C"+fmt.Sprintf("%d", idx+2))
			report.MergeCell(sheet, "F"+fmt.Sprintf("%d", idx+2), "G"+fmt.Sprintf("%d", idx+2))
			report.MergeCell(sheet, "J"+fmt.Sprintf("%d", idx+2), "K"+fmt.Sprintf("%d", idx+2))
			report.SetCellStyle(sheet, "B"+fmt.Sprintf("%d", idx+2), "B"+fmt.Sprintf("%d", idx+2), styleFutureValueMoney)
			report.SetCellStyle(sheet, "F"+fmt.Sprintf("%d", idx+2), "F"+fmt.Sprintf("%d", idx+2), styleFutureValueMoney)
			report.SetCellStyle(sheet, "J"+fmt.Sprintf("%d", idx+2), "J"+fmt.Sprintf("%d", idx+2), styleFutureValueMoney)

			report.SetCellValue(sheet, "B"+fmt.Sprintf("%d", idx), "Due to call centre load...")
			report.SetCellValue(sheet, "B"+fmt.Sprintf("%d", idx+1), fmt.Sprintf("%.0f", details["lost_agent_hours_14d"].(float64))+" Hours")
			report.SetCellValue(sheet, "F"+fmt.Sprintf("%d", idx+1), fmt.Sprintf("%.0f", details["lost_agent_hours_21d"].(float64))+" Hours")
			report.SetCellValue(sheet, "J"+fmt.Sprintf("%d", idx+1), fmt.Sprintf("%.0f", details["lost_agent_hours_28d"].(float64))+" Hours")
			report.SetCellValue(sheet, "B"+fmt.Sprintf("%d", idx+2), fmt.Sprintf("£%.0f", details["hours_lost_cost_14d"].(float64)))
			report.SetCellValue(sheet, "F"+fmt.Sprintf("%d", idx+2), fmt.Sprintf("£%.0f", details["hours_lost_cost_21d"].(float64)))
			report.SetCellValue(sheet, "J"+fmt.Sprintf("%d", idx+2), fmt.Sprintf("£%.0f", details["hours_lost_cost_28d"].(float64)))
			idx += 4
		} else if useCase == "incurred_costs" {
			report.SetCellValue(sheet, "B"+fmt.Sprintf("%d", idx), "Due to incurred costs...")
			report.SetCellValue(sheet, "B"+fmt.Sprintf("%d", idx+1), fmt.Sprintf("£%d", details["costs_incurred_14d"].(int)))
			report.SetCellValue(sheet, "F"+fmt.Sprintf("%d", idx+1), fmt.Sprintf("£%d", details["costs_incurred_21d"].(int)))
			report.SetCellValue(sheet, "J"+fmt.Sprintf("%d", idx+1), fmt.Sprintf("£%d", details["costs_incurred_28d"].(int)))
			idx += 3
		}
	}
}

func addUserBreakdownChart(sheet string, report *excelize.File, data interface{}, posX string) {
	chartData := data.([3]interface{})
	labels := []string{}
	values := []string{}

	for _, i := range chartData {
		item := i.([2]interface{})
		labels = append(labels, item[0].(string))
		values = append(values, fmt.Sprintf("%d", item[1].(int)))
	}

	cats := strings.Join(labels, `\",\"`)
	vals := strings.Join(values, ", ")

	if err := report.AddChart(sheet, posX, `{
		"type": "pie",
		"series": [
			{
				"name": "Breakdown",
				"categories": "{\"`+cats+`\"}",
				"values": "{`+vals+`}"
			}
		],
		"legend": {
			"position": "right"
		},
		"title": {
			"name": "Lost user breakdown per channel"
		},
		"plotarea": {
			"show_bubble_size": true,
			"show_cat_name": false,
            "show_leader_lines": false,
            "show_percent": true,
            "show_series_name": false,
            "show_val": false
		},
		"chartarea": {
			"border": {
				"none": true
			}
		},
		"dimension": {
			"height": 220,
			"width": 460
		}
	}`); err != nil {
		util.FailOnError(err, "error adding chart")
	}
}

func addDailyBreakdownChart(sheet string, report *excelize.File, data interface{}, posX string) {
	chartData := data.([]interface{})
	labels := []string{}
	values := []string{}

	for _, i := range chartData {
		item := i.([]interface{})
		labels = append(labels, item[0].(string))
		values = append(values, fmt.Sprintf("%d", item[1].(int)))
	}

	cats := strings.Join(labels, `\",\"`)
	vals := strings.Join(values, ", ")

	if err := report.AddChart(sheet, posX, `{
		"type": "col",
		"series": [
			{
				"name": "Breakdown",
				"categories": "{\"`+cats+`\"}",
				"values": "{`+vals+`}"
			}
		],
		"legend": {
			"none": true
		},
		"title": {
			"name": "Occurrence of error over time"
		},
		"plotarea": {
			"show_bubble_size": true,
			"show_cat_name": false,
            "show_leader_lines": false,
            "show_percent": false,
            "show_series_name": false,
            "show_val": true
		},
		"chartarea": {
			"border": {
				"none": true
			}
		},
		"dimension": {
			"height": 220,
			"width": 690
		}
	}`); err != nil {
		util.FailOnError(err, "error adding chart")
	}
}

func populateSummarySheet(sheet string, report *excelize.File, reportData map[string]map[string]interface{}, env environment.Environment, config config.Config) {
	styleSubtitle := getExcelStyle("subtitle", report)
	styleLogoBump := getExcelStyle("logoBump", report)
	styleSubtitle2 := getExcelStyle("subtitle2", report)
	styleSummaryDetail := getExcelStyle("summaryDetail", report)
	styleSummaryMoney := getExcelStyle("summaryMoney", report)
	styleLink := getExcelStyle("link", report)

	// Layout
	report.SetRowHeight(sheet, 3, 57)
	report.SetColWidth(sheet, "A", "A", 3)
	report.SetColWidth(sheet, "B", "B", 23)
	report.SetColWidth(sheet, "C", "C", 3)
	report.SetColWidth(sheet, "D", "D", 14)
	report.MergeCell(sheet, "B3", "D3")
	report.MergeCell(sheet, "B9", "C9")
	report.MergeCell(sheet, "E3", "I3")
	report.MergeCell(sheet, "B11", "D11")
	report.MergeCell(sheet, "E11", "G11")
	report.SetCellStyle(sheet, "E3", "E3", styleLogoBump)
	report.SetCellStyle(sheet, "B6", "B10", styleSubtitle)
	report.SetCellStyle(sheet, "D6", "D6", styleSummaryDetail)
	report.SetCellStyle(sheet, "D7", "D7", styleLink)
	report.SetCellStyle(sheet, "D8", "D8", styleSummaryDetail)
	report.SetCellStyle(sheet, "B11", "B11", styleSummaryDetail)
	report.SetCellStyle(sheet, "E11", "E11", styleSubtitle2)

	// Images
	if dtLogo, err := imgs.ReadFile("img/derran_logo.png"); err == nil {
		report.AddPictureFromBytes(sheet, "B3", `{
			"lock_aspect_ratio": false,
			"positioning": "oneCell",
			"x_scale": 0.6,
			"y_scale": 0.6,
			"x_offset": 6,
			"y_offset": 10
		}`, "dt_logo", ".png", dtLogo)
	}

	// Flat text
	//report.SetCellValue(sheet, "E3", "error analyser")
	report.SetCellValue(sheet, "B6", "Environment name")
	report.SetCellValue(sheet, "D6", env.GetName())
	report.SetCellValue(sheet, "B7", "Environment URL")
	report.SetCellValue(sheet, "D7", env.GetEnvironmentUrl())
	report.SetCellValue(sheet, "B8", "Configuration")
	report.SetCellValue(sheet, "D8", config.GetName())
	report.SetCellValue(sheet, "B11", "Error name")
	report.SetCellValue(sheet, "E11", "Monetary impact")

	// Errors analysed by monetary impact
	errors := make(map[string]int)
	for envErr, details := range reportData {
		errors[envErr] = details["total_impact"].(int)
	}
	report.SetCellValue(sheet, "B10", fmt.Sprintf("%d", len(errors))+" errors analysed...")
	i := 12
	for _, k := range util.SortMapDesc(errors) {
		idx := fmt.Sprintf("%d", i)
		report.MergeCell(sheet, "B"+idx, "D"+idx)
		report.MergeCell(sheet, "E"+idx, "G"+idx)
		report.SetCellStyle(sheet, "B"+idx, "B"+idx, styleLink)
		report.SetCellStyle(sheet, "E"+idx, "E"+idx, styleSummaryMoney)

		report.SetCellValue(sheet, "B"+idx, k.Key)
		report.SetCellHyperLink(sheet, "B"+idx, "'"+k.Key+"'!A1", "Location")
		report.SetCellValue(sheet, "E"+idx, fmt.Sprintf("£%d", k.Value))

		i++
	}
}
