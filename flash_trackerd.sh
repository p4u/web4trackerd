esptool.py --chip esp32 --port "/dev/ttyACM0" --baud 921600 write_flash --compress --flash_mode dio --flash_size detect 0x10000 EU868.bin
