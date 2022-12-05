package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-sensors/core/gas"
	"github.com/go-sensors/core/humidity"
	"github.com/go-sensors/core/i2c"
	"github.com/go-sensors/core/temperature"
	"github.com/go-sensors/core/units"
	"github.com/go-sensors/rpi-sensironscd30-config/internal/log"
	"github.com/go-sensors/rpii2c"
	"github.com/go-sensors/sensironscd30"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/syncromatics/go-kit/v2/cmd"
)

var (
	rootCmd = &cobra.Command{
		Use:   "rpi-sensironscd30-config",
		Short: "Configure a Sensiron SCD30 gas sensor",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logLevel, err := cmd.Flags().GetString("log-level")
			if err != nil {
				logLevel = "warn"
			}

			isTerminal := isatty.IsTerminal(os.Stdout.Fd())
			log.InitializeLogger(isTerminal, logLevel)
			cmd.SilenceErrors = !isTerminal
			cmd.SilenceUsage = !isTerminal
		},
	}
	temperatureOffset = &cobra.Command{
		Use:   "temperature-offset [optional offset °C]",
		Short: "Get and set temperature offset",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			bus, err := cobraCmd.Flags().GetInt("sensironscd30-i2c-bus")
			if err != nil {
				return err
			}

			address, err := cobraCmd.Flags().GetUint8("sensironscd30-i2c-addr")
			if err != nil {
				return err
			}

			i2cPortFactory, err := rpii2c.NewI2CPort(bus, &i2c.I2CPortConfig{Address: byte(address)})
			if err != nil {
				return errors.Wrap(err, "failed to initialize I2C port factory")
			}

			sensor := sensironscd30.NewSensor(i2cPortFactory)
			group := cmd.NewProcessGroup(context.Background())
			group.Go(func() error {
				if len(args) > 1 {
					return errors.New("invalid number of arguments")
				}

				temperatureOffset, err := sensor.GetTemperatureOffset(group.Context())
				if err != nil {
					return err
				}

				fmt.Printf("Current temperature offset: %v\n", temperatureOffset.String())

				if len(args) == 1 {
					degreesCelsius, err := strconv.ParseFloat(args[0], 64)
					if err != nil {
						return err
					}

					temperatureOffset := units.Temperature(degreesCelsius * float64(units.DegreeCelsius))
					err = sensor.SetTemperatureOffset(group.Context(), temperatureOffset)
					if err != nil {
						return err
					}

					fmt.Printf("New temperature offset:     %v\n", temperatureOffset.String())
				}

				return nil
			})

			return group.Wait()
		},
	}
	frc = &cobra.Command{
		Use:   "frc [CO2 ppm]",
		Short: "Set a new baseline CO2 concentration value (forced recalibration)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			bus, err := cobraCmd.Flags().GetInt("sensironscd30-i2c-bus")
			if err != nil {
				return err
			}

			address, err := cobraCmd.Flags().GetUint8("sensironscd30-i2c-addr")
			if err != nil {
				return err
			}

			i2cPortFactory, err := rpii2c.NewI2CPort(bus, &i2c.I2CPortConfig{Address: byte(address)})
			if err != nil {
				return errors.Wrap(err, "failed to initialize I2C port factory")
			}

			partsPerMillion, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return err
			}

			totalRuntime := 3 * time.Minute
			forceRecalibrationAfter := 2 * time.Minute
			baselineConcentration := units.Concentration(partsPerMillion * float64(units.PartPerMillion))
			sensor := sensironscd30.NewSensor(i2cPortFactory,
				sensironscd30.WithForcedRecalibrationValue(baselineConcentration, forceRecalibrationAfter))

			handler := &stdoutHandler{}
			gasConsumer := gas.NewConsumer(sensor, handler)
			temperatureConsumer := temperature.NewConsumer(sensor, handler)
			humidityConsumer := humidity.NewConsumer(sensor, handler)

			fmt.Printf("Running forced recalibration for a total duration of %v\n", totalRuntime)
			ctx, cancel := context.WithTimeout(context.Background(), totalRuntime)
			defer cancel()

			group := cmd.NewProcessGroup(ctx)
			group.Start(gasConsumer.Run)
			group.Start(temperatureConsumer.Run)
			group.Start(humidityConsumer.Run)
			group.Start(sensor.Run)

			err = group.Wait()
			if err != nil {
				return err
			}

			fmt.Println("Completed forced recalibration")
			return nil
		},
	}
)

type stdoutHandler struct{}

func (*stdoutHandler) HandleGasConcentration(_ context.Context, value *gas.Concentration) error {
	fmt.Printf("Gas concentration: %v %v\n", value.Gas, value.Amount.String())
	return nil
}

func (*stdoutHandler) HandleTemperature(_ context.Context, value *units.Temperature) error {
	fmt.Printf("Temperature: %v\n", value.String())
	return nil
}

func (*stdoutHandler) HandleRelativeHumidity(_ context.Context, value *units.RelativeHumidity) error {
	fmt.Printf("Relative humidity: %v\n", value.String())
	return nil
}

func init() {
	rootCmd.PersistentFlags().String("log-level", "warn", "Determines the amount of detail included in the log output; valid options are: fatal, error, warn, info, debug")
	rootCmd.PersistentFlags().Int("sensironscd30-i2c-bus", 1, "Number of I2C bus on which to find the sensor")
	rootCmd.PersistentFlags().Uint8("sensironscd30-i2c-addr", sensironscd30.GetDefaultI2CPortConfig().Address, "7-bit I2C address of the sensor")
	rootCmd.AddCommand(temperatureOffset)
	rootCmd.AddCommand(frc)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal("failed to terminate cleanly",
			"err", err)
	}
}
