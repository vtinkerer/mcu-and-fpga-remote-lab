import sys 
from time import sleep
import re
import os
import fcntl
from WF_SDK import device, error, static

from ctypes import *
from dwfconstants import *
import sys

if sys.platform.startswith("win"):
    dwf = cdll.dwf
elif sys.platform.startswith("darwin"):
    dwf = cdll.LoadLibrary("/Library/Frameworks/dwf.framework/dwf")
else:
    dwf = cdll.LoadLibrary("libdwf.so")

delimiter = '\n'

def write_to_stdout(message):
    sys.stdout.write(message + delimiter)
    sys.stdout.flush()

def parse_command(command):
    pin_set_match = re.match(r"set=(\d+):(\d+)", command)
    if pin_set_match:
        return SetPinCommand(int(pin_set_match.group(1)), int(pin_set_match.group(2)))
    pin_read_match = re.match(r"read=(\d+)", command)
    if pin_read_match:
        return ReadPinCommand(int(pin_read_match.group(1)))
    pin_click_match = re.match(r"click=(\d+)", command)
    if pin_click_match:
        return ClickPinCommand(int(pin_click_match.group(1)))
    return None

class SetPinCommand:
    def __init__(self, pin, value):
        self.pin = pin
        self.value = value

    def handle(self, device_data):
        write_to_stdout(f"Setting pin {self.pin} to {self.value}")
        static.set_state(device_data=device_data, channel=self.pin, value=self.value == 1)

class ClickPinCommand:
    def __init__(self, pin):
        self.pin = pin

    def handle(self):
        write_to_stdout(f"Clicking pin {self.pin} DOWN")
        static.set_state(device_data=device_data, channel=self.pin, value=False)
        sleep(0.1)
        write_to_stdout(f"Clicking pin {self.pin} UP")
        static.set_state(device_data=device_data, channel=self.pin, value=True)

class ReadPinCommand:
    def __init__(self, pin):
        self.pin = pin

    def handle(self):
        write_to_stdout(f"Reading pin {self.pin}")

fd = sys.stdin.fileno()
fl = fcntl.fcntl(fd, fcntl.F_GETFL)
fcntl.fcntl(fd, fcntl.F_SETFL, fl | os.O_NONBLOCK)

def read_input():
    input_commands = []
    while True:
        try:
            line = sys.stdin.readline()
            if not line:  # EOF
                break
            input_commands.append(line.strip())
        except IOError:
            break
    return input_commands

try:
    device_data = device.open()
    static.set_mode(device_data=device_data, channel=0, output=True)
    static.set_mode(device_data=device_data, channel=1, output=True)
    static.set_mode(device_data=device_data, channel=2, output=True)

    write_to_stdout("Ready")

    static.set_state(device_data=device_data, channel=0, value=True)
    static.set_state(device_data=device_data, channel=1, value=True)
    static.set_state(device_data=device_data, channel=2, value=True)

    while True:
        input_commands = read_input()
        for command in input_commands:
            parsed_command = parse_command(command)
            if parsed_command:
                parsed_command.handle(device_data)
            else:
                write_to_stdout("Invalid command")
                write_to_stdout(str(input_commands))
        if not input_commands:
            sleep(0.1)

except error as e:
    sys.stderr.write(str(e))
    sys.stderr.flush()
    # dwf.FDwfDeviceClose(hdwf)
    device.close(device_data)

    