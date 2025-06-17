package prusalink

import (
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/pstrobl96/prusa_exporter/config"
	"github.com/rs/zerolog/log"
)

// Collector is a struct of all printer metrics
type Collector struct {
	printerTemp               *prometheus.Desc
	printerTempTarget         *prometheus.Desc
	printerPrintTime          *prometheus.Desc
	printerPrintTimeRemaining *prometheus.Desc
	printerPrintProgressRatio *prometheus.Desc
	printerFiles              *prometheus.Desc
	printerMaterial           *prometheus.Desc
	printerUp                 *prometheus.Desc
	printerNozzleSize         *prometheus.Desc
	printerStatus             *prometheus.Desc
	printerAxis               *prometheus.Desc
	printerFlow               *prometheus.Desc
	printerInfo               *prometheus.Desc
	printerMMU                *prometheus.Desc
	printerFanSpeedRpm        *prometheus.Desc
	printerPrintSpeedRatio    *prometheus.Desc
	printerJobImage           *prometheus.Desc

	configuration config.Config
	commonLabels  []string
}

// NewCollector returns a new Collector for printer metrics
func NewCollector(config config.Config) *Collector {
	configuration = config
	defaultLabels := []string{"printer_address", "printer_model", "printer_name", "printer_job_name", "printer_job_path"}
	return &Collector{
		configuration: config,
		commonLabels:  defaultLabels,

		printerTemp:               prometheus.NewDesc("prusa_temperature_celsius", "Current temp of printer in Celsius", append(defaultLabels, "printer_heated_element"), nil),
		printerTempTarget:         prometheus.NewDesc("prusa_temperature_target_celsius", "Target temp of printer in Celsius", append(defaultLabels, "printer_heated_element"), nil),
		printerPrintTimeRemaining: prometheus.NewDesc("prusa_printing_time_remaining_seconds", "Returns time that remains for completion of current print", defaultLabels, nil),
		printerPrintProgressRatio: prometheus.NewDesc("prusa_printing_progress_ratio", "Returns information about completion of current print in ratio (0.0-1.0)", defaultLabels, nil),
		printerFiles:              prometheus.NewDesc("prusa_files_count", "Number of files in storage", append(defaultLabels, "printer_storage"), nil),
		printerMaterial:           prometheus.NewDesc("prusa_material_info", "Returns information about loaded filament. Returns 0 if there is no loaded filament", append(defaultLabels, "printer_filament"), nil),
		printerPrintTime:          prometheus.NewDesc("prusa_print_time_seconds", "Returns information about current print time.", defaultLabels, nil),
		printerUp:                 prometheus.NewDesc("prusa_up", "Return information about online printers. If printer is registered as offline then returned value is 0.", []string{"printer_address", "printer_model", "printer_name"}, nil),
		printerNozzleSize:         prometheus.NewDesc("prusa_nozzle_size_meters", "Returns information about selected nozzle size.", defaultLabels, nil),
		printerStatus:             prometheus.NewDesc("prusa_status_info", "Returns information status of printer.", append(defaultLabels, "printer_state"), nil),
		printerAxis:               prometheus.NewDesc("prusa_axis", "Returns information about position of axis.", append(defaultLabels, "printer_axis"), nil),
		printerFlow:               prometheus.NewDesc("prusa_print_flow_ratio", "Returns information about of filament flow in ratio (0.0 - 1.0).", defaultLabels, nil),
		printerInfo:               prometheus.NewDesc("prusa_info", "Returns information about printer.", append(defaultLabels, "api_version", "server_version", "version_text", "prusalink_name", "printer_location", "serial_number", "printer_hostname"), nil),
		printerMMU:                prometheus.NewDesc("prusa_mmu", "Returns information if MMU is enabled.", defaultLabels, nil),
		printerFanSpeedRpm:        prometheus.NewDesc("prusa_fan_speed_rpm", "Returns information about speed of hotend fan in rpm.", append(defaultLabels, "fan"), nil),
		printerPrintSpeedRatio:    prometheus.NewDesc("prusa_print_speed_ratio", "Current setting of printer speed in values from 0.0 - 1.0", defaultLabels, nil),
		printerJobImage:           prometheus.NewDesc("prusa_job_image", "Returns information about image of current print job.", append(defaultLabels, "printer_job_image"), nil),
	}
}

// Describe implements prometheus.Collector
func (collector *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.printerTemp
	ch <- collector.printerTempTarget
	ch <- collector.printerFiles
	ch <- collector.printerPrintTime
	ch <- collector.printerPrintTimeRemaining
	ch <- collector.printerPrintProgressRatio
	ch <- collector.printerPrintSpeedRatio
	ch <- collector.printerMaterial
	ch <- collector.printerUp
	ch <- collector.printerNozzleSize
	ch <- collector.printerStatus
	ch <- collector.printerAxis
	ch <- collector.printerFlow
	ch <- collector.printerInfo
	ch <- collector.printerMMU
	ch <- collector.printerFanSpeedRpm
	ch <- collector.printerJobImage
}

// Collect implements prometheus.Collector
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	for _, s := range c.configuration.Printers {
		wg.Add(1)
		go func(s config.Printers) {
			defer wg.Done()

			log.Debug().Msg("Printer scraping at " + s.Address)
			printerUp := prometheus.MustNewConstMetric(c.printerUp, prometheus.GaugeValue,
				0, s.Address, s.Type, s.Name)

			job, err := GetJob(s)
			if err != nil {
				log.Error().Msg("Error while scraping job endpoint at " + s.Address + " - " + err.Error())
				ch <- printerUp
				return
			}

			printer, err := GetPrinter(s)
			if err != nil {
				log.Error().Msg("Error while scraping printer endpoint at " + s.Address + " - " + err.Error())
				ch <- printerUp
				return
			}

			version, err := GetVersion(s)
			if err != nil {
				log.Error().Msg("Error while scraping version endpoint at " + s.Address + " - " + err.Error())
				ch <- printerUp
				return
			}

			status, err := GetStatus(s)

			if err != nil {
				log.Error().Msg("Error while scraping status endpoint at " + s.Address + " - " + err.Error())
			}

			info, err := GetInfo(s)

			if err != nil {
				log.Error().Msg("Error while scraping info endpoint at " + s.Address + " - " + err.Error())
			}

			printerInfo := prometheus.MustNewConstMetric(
				c.printerInfo, prometheus.GaugeValue,
				1,
				c.GetLabels(s, job, version.API, version.Server, version.Text, info.Name, info.Location, info.Serial, info.Hostname)...)

			ch <- printerInfo

			printerFanHotend := prometheus.MustNewConstMetric(c.printerFanSpeedRpm, prometheus.GaugeValue,
				status.Printer.FanHotend, c.GetLabels(s, job, "hotend")...)

			ch <- printerFanHotend

			printerFanPrint := prometheus.MustNewConstMetric(c.printerFanSpeedRpm, prometheus.GaugeValue,
				status.Printer.FanPrint, c.GetLabels(s, job, "print")...)

			ch <- printerFanPrint

			printerNozzleSize := prometheus.MustNewConstMetric(c.printerNozzleSize, prometheus.GaugeValue,
				info.NozzleDiameter, c.GetLabels(s, job)...)

			ch <- printerNozzleSize

			printSpeed := prometheus.MustNewConstMetric(
				c.printerPrintSpeedRatio, prometheus.GaugeValue,
				printer.Telemetry.PrintSpeed/100,
				c.GetLabels(s, job)...)

			ch <- printSpeed

			printTime := prometheus.MustNewConstMetric(
				c.printerPrintTime, prometheus.GaugeValue,
				job.Progress.PrintTime,
				c.GetLabels(s, job)...)

			ch <- printTime

			printTimeRemaining := prometheus.MustNewConstMetric(
				c.printerPrintTimeRemaining, prometheus.GaugeValue,
				job.Progress.PrintTimeLeft,
				c.GetLabels(s, job)...)

			ch <- printTimeRemaining

			printProgress := prometheus.MustNewConstMetric(
				c.printerPrintProgressRatio, prometheus.GaugeValue,
				job.Progress.Completion,
				c.GetLabels(s, job)...)

			ch <- printProgress

			material := prometheus.MustNewConstMetric(
				c.printerMaterial, prometheus.GaugeValue,
				BoolToFloat(!(strings.Contains(printer.Telemetry.Material, "-"))),
				c.GetLabels(s, job, printer.Telemetry.Material)...)

			ch <- material

			printerAxisX := prometheus.MustNewConstMetric(
				c.printerAxis, prometheus.GaugeValue,
				printer.Telemetry.AxisX,
				c.GetLabels(s, job, "x")...)

			ch <- printerAxisX

			printerAxisY := prometheus.MustNewConstMetric(
				c.printerAxis, prometheus.GaugeValue,
				printer.Telemetry.AxisY,
				c.GetLabels(s, job, "y")...)

			ch <- printerAxisY

			printerAxisZ := prometheus.MustNewConstMetric(
				c.printerAxis, prometheus.GaugeValue,
				printer.Telemetry.AxisZ,
				c.GetLabels(s, job, "z")...)

			ch <- printerAxisZ

			printerFlow := prometheus.MustNewConstMetric(c.printerFlow, prometheus.GaugeValue,
				status.Printer.Flow/100, c.GetLabels(s, job)...)

			ch <- printerFlow

			printerMMU := prometheus.MustNewConstMetric(c.printerMMU, prometheus.GaugeValue,
				BoolToFloat(info.Mmu), c.GetLabels(s, job)...)
			ch <- printerMMU

			printerBedTemp := prometheus.MustNewConstMetric(c.printerTemp, prometheus.GaugeValue,
				printer.Temperature.Bed.Actual, c.GetLabels(s, job, "bed")...)

			ch <- printerBedTemp

			printerBedTempTarget := prometheus.MustNewConstMetric(c.printerTempTarget, prometheus.GaugeValue,
				printer.Temperature.Bed.Target, c.GetLabels(s, job, "bed")...)

			ch <- printerBedTempTarget

			printerToolTempTarget := prometheus.MustNewConstMetric(c.printerTempTarget, prometheus.GaugeValue,
				printer.Temperature.Tool0.Target, c.GetLabels(s, job, "tool0")...)

			ch <- printerToolTempTarget

			printerToolTemp := prometheus.MustNewConstMetric(c.printerTemp, prometheus.GaugeValue,
				printer.Temperature.Tool0.Actual, c.GetLabels(s, job, "tool0")...)

			ch <- printerToolTemp

			printerStatus := prometheus.MustNewConstMetric(
				c.printerStatus, prometheus.GaugeValue,
				getStateFlag(printer),
				c.GetLabels(s, job, printer.State.Text)...)

			ch <- printerStatus

			if getStateFlag(printer) == 4 {
				image, err := GetJobImage(s, job.Job.File.Path)

				if err != nil {
					log.Error().Msg("Error while scraping image endpoint at " + s.Address + " - " + err.Error())
				} else {
					printerJobImage := prometheus.MustNewConstMetric(c.printerJobImage, prometheus.GaugeValue,
						1, c.GetLabels(s, job, image)...)

					ch <- printerJobImage
				}

			}

			printerUp = prometheus.MustNewConstMetric(c.printerUp, prometheus.GaugeValue,
				1, s.Address, s.Type, s.Name)

			ch <- printerUp

			log.Debug().Msg("Scraping done at " + s.Address)
		}(s)
	}
	wg.Wait()
}

// GetLabels is used to get the labels for the given printer and job
func (c *Collector) GetLabels(printer config.Printers, job Job, labelValues ...string) []string {
	commonValues := make([]string, len(c.commonLabels), len(c.commonLabels)+len(labelValues))

	for i, l := range c.commonLabels {
		switch l {
		case "printer_address":
			commonValues[i] = printer.Address
		case "printer_model":
			commonValues[i] = printer.Type
		case "printer_name":
			commonValues[i] = printer.Name

		// job is passed by value, and none of the fields are pointers,
		// so we don't need to worry about nil dereferences.
		case "printer_job_name":
			commonValues[i] = job.Job.File.Name
		case "printer_job_path":
			commonValues[i] = job.Job.File.Path
		}
	}
	return append(commonValues, labelValues...)
}
