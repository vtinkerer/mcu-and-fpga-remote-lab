# MCU and FPGA Remote Lab (back-end)
Repository for the back-end part of MCU and FPGA remote laboratory for programming of MCU and FPGA boards, interaction with digital inputs, work with functional generator and capturing signals via oscilloscope

## Used technologies
Go, WaveForms SDK API, purego, gin, Docker.

## How to run back-end
1. Clone the repository.
2. Download and install [Go](https://go.dev/doc/install).
3. Ensure that purego library and gin framework are installed.
4. Connect all the necessary Analog Discovery 2 equipment to the device, where back-end will be run.
5. In the root folder of the project execute `go run .` command.
6. If the back-end is configured to be pm process, then any update can be applied with the use of `pm2 restart 0` command.