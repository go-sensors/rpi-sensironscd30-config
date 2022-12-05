# go-sensors/rpi-sensironscd30-config

A utility to configure a [Sensiron SCD30][sensironscd30] gas sensor for detecting carbon dioxide concentration, temperature, and humidity.

[sensironscd30]: https://sensirion.com/us/products/catalog/SCD30/

## Quickstart

Get the sensor's current temperature offset:

```bash
rpi-sensironscd30-config temperature-offset
```

Set a new temperature offset:

```bash
rpi-sensironscd30-config temperature-offset 1.2
```

Perform a forced recalibration (FRC) routine. Per vendor specifications, the sensor must be in an environment with a consistent concentration of CO2 for at least 2 minutes.

The new baseline may be obtained from another calibrated source (other sensor, etc.), or may be approximated by leaving the sensor outside, away from plants, humans, etc. and using a value between 400 and 420 ppm.

```bash
rpi-sensironscd30-config frc 420
```

## Building

TBD

## Code of Conduct

We are committed to fostering an open and welcoming environment. Please read our [code of conduct](CODE_OF_CONDUCT.md) before participating in or contributing to this project.

## Contributing

We welcome contributions and collaboration on this project. Please read our [contributor's guide](CONTRIBUTING.md) to understand how best to work with us.

## License and Authors

[![Daniel James logo](https://secure.gravatar.com/avatar/eaeac922b9f3cc9fd18cb9629b9e79f6.png?size=16) Daniel James](https://github.com/thzinc)

[![license](https://img.shields.io/github/license/go-sensors/rpi-sensironscd30-config.svg)](https://github.com/go-sensors/rpi-sensironscd30-config/blob/master/LICENSE)
[![GitHub contributors](https://img.shields.io/github/contributors/go-sensors/rpi-sensironscd30-config.svg)](https://github.com/go-sensors/rpi-sensironscd30-config/graphs/contributors)

This software is made available by Daniel James under the MIT license.
